package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"crm_backend/internal/model"
	"crm_backend/pkg/events"
	"crm_backend/pkg/telegram"
)

// TelegramBotService реализует сценарий приёмной комиссии:
//   greet → ask_name → ask_program → ask_phone → manager (handoff)
// Каждое входящее/исходящее сообщение сохраняется как Interaction
// с direction=inbound|outbound, type=messenger.
// После handoff бот молчит, сообщения студента просто пишутся в историю.

type TelegramChatRepo interface {
	GetByChatID(ctx context.Context, chatID int64) (*model.TelegramChat, error)
	GetByLeadID(ctx context.Context, leadID int) (*model.TelegramChat, error)
	Create(ctx context.Context, chatID int64, username, firstName string) (*model.TelegramChat, error)
	UpdateState(ctx context.Context, id int, state string) error
	SetBotActive(ctx context.Context, id int, active bool) error
	SetLeadID(ctx context.Context, id int, leadID int) error
	SetCollected(ctx context.Context, id int, field, value string) error
}

type ProgramLookup interface {
	FindIDByName(ctx context.Context, name string) (int, bool, error)
}

type ProgramListSource interface {
	ActiveNames(ctx context.Context) ([]string, error)
}

type BotSettingsSource interface {
	Get(ctx context.Context, key, fallback string) string
}

type TelegramBotService struct {
	client       *telegram.Client
	chats        TelegramChatRepo
	leadSvc      ILeadService
	intSvc       IInteractionService
	contactRepo  ContactRepository
	leadRepo     LeadRepository
	programRepo  ProgramLookup
	programList  ProgramListSource
	botSettings  BotSettingsSource
	bus          *events.Bus
}

func NewTelegramBotService(
	token string,
	chats TelegramChatRepo,
	leadSvc ILeadService,
	intSvc IInteractionService,
	contactRepo ContactRepository,
	leadRepo LeadRepository,
	programRepo ProgramLookup,
	programList ProgramListSource,
	botSettings BotSettingsSource,
	bus *events.Bus,
) *TelegramBotService {
	var client *telegram.Client
	if token != "" {
		client = telegram.NewClient(token)
	}
	return &TelegramBotService{
		client: client, chats: chats,
		leadSvc: leadSvc, intSvc: intSvc,
		contactRepo: contactRepo, leadRepo: leadRepo,
		programRepo: programRepo,
		programList: programList,
		botSettings: botSettings,
		bus:         bus,
	}
}

func (s *TelegramBotService) publishMessage(leadID *int, direction, content string) {
	if s.bus == nil || leadID == nil {
		return
	}
	s.bus.Publish(events.Event{
		Type: "telegram.message",
		Payload: map[string]interface{}{
			"lead_id":   *leadID,
			"direction": direction,
			"content":   content,
		},
	})
}

// HasClient — true, если задан bot token и исходящая отправка возможна.
func (s *TelegramBotService) HasClient() bool { return s.client != nil }

// fallbackPrograms используется только если БД недоступна.
var fallbackPrograms = []string{
	"Professional Certificate in English (PCIE)",
	"BSc (Hons) Cyber Security",
	"BSc (Hons) Information Technology",
	"BA (Hons) Business and Financial Management",
	"BA (Hons) Business and Marketing Management",
	"BSc (Hons) International Tourism and Hospitality Management (Top-Up)",
	"Masters of Business Administration (MBA)",
}

func (s *TelegramBotService) loadPrograms(ctx context.Context) []string {
	if s.programList != nil {
		if names, err := s.programList.ActiveNames(ctx); err == nil && len(names) > 0 {
			return names
		}
	}
	return fallbackPrograms
}

func (s *TelegramBotService) setting(ctx context.Context, key, fallback string) string {
	if s.botSettings == nil {
		return fallback
	}
	return s.botSettings.Get(ctx, key, fallback)
}

func (s *TelegramBotService) greetingText(ctx context.Context, firstName string) string {
	name := firstName
	if name == "" {
		name = "абитуриент"
	}
	tpl := s.setting(ctx, "greeting",
		"Здравствуйте, {name}! 👋\n\nЯ бот приёмной комиссии MDIS Tashkent. Чем могу помочь?")
	return strings.ReplaceAll(tpl, "{name}", name)
}

// menuKeyboard — главное меню с FAQ + «Подать заявку».
func menuKeyboard() [][]telegram.InlineButton {
	return [][]telegram.InlineButton{
		{{Text: "Узнать условия поступления", CallbackData: "menu:admission"}},
		{{Text: "Регистрация на мероприятие", CallbackData: "menu:event"}},
		{{Text: "Записаться на EET тест", CallbackData: "menu:eet"}},
		{{Text: "📝 Подать заявку", CallbackData: "menu:apply"}},
	}
}

// programKeyboard собирает inline-клавиатуру с программами.
// callback_data = "program:<index>" — индекс в локальном массиве.
func programKeyboard(programs []string) [][]telegram.InlineButton {
	rows := make([][]telegram.InlineButton, 0, len(programs))
	for i, p := range programs {
		rows = append(rows, []telegram.InlineButton{{
			Text:         p,
			CallbackData: fmt.Sprintf("program:%d", i),
		}})
	}
	return rows
}

func parseProgram(text string, programs []string) string {
	t := strings.TrimSpace(text)
	for i, p := range programs {
		idx := fmt.Sprintf("%d", i+1)
		if t == idx {
			return p
		}
	}
	lt := strings.ToLower(t)
	for _, p := range programs {
		if strings.Contains(strings.ToLower(p), lt) || strings.Contains(lt, strings.ToLower(p)) {
			return p
		}
	}
	return ""
}

// HandleIncoming обрабатывает входящее сообщение от Telegram-вебхука.
// Сохраняет inbound-interaction, прогоняет state machine, отправляет ответ.
func (s *TelegramBotService) HandleIncoming(ctx context.Context, payload *model.TelegramWebhookRequest) error {
	// Клик inline-кнопки приходит отдельным апдейтом callback_query.
	if payload.CallbackQuery != nil {
		return s.handleCallback(ctx, payload.CallbackQuery)
	}
	if payload.Message.Text == "" || payload.Message.Chat.ID == 0 {
		return nil
	}
	chatID := int64(payload.Message.Chat.ID)
	text := strings.TrimSpace(payload.Message.Text)
	firstName := payload.Message.From.FirstName
	username := payload.Message.From.Username

	chat, err := s.chats.GetByChatID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat == nil {
		chat, err = s.chats.Create(ctx, chatID, username, firstName)
		if err != nil {
			return err
		}
	}

	// Сохраняем inbound сообщение (если уже есть лид — привяжем)
	if chat.LeadID != nil {
		_, _ = s.intSvc.AddInteraction(ctx, &model.CreateInteractionRequest{
			LeadID:    chat.LeadID,
			Type:      "messenger",
			Direction: "inbound",
			Content:   text,
		})
		s.publishMessage(chat.LeadID, "inbound", text)
	}

	// Если менеджер уже перехватил — бот молчит
	if !chat.BotActive || chat.BotState == model.BotStateManager {
		return nil
	}

	// Команда /start всегда сбрасывает в начало (показ главного меню).
	if text == "/start" {
		_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateMenu)
		return s.replyWithKeyboard(ctx, chat, s.greetingText(ctx, firstName), menuKeyboard())
	}

	askProgramText := s.setting(ctx, "ask_program",
		"Спасибо! Какая программа вас интересует? Выберите вариант на кнопке ниже 👇")
	askPhoneText := s.setting(ctx, "ask_phone",
		"Отлично! Оставьте, пожалуйста, ваш номер телефона для связи (например, +992 90 123 45 67).")
	handoffText := s.setting(ctx, "handoff",
		"Спасибо! Ваша заявка принята. ✅\nМенеджер свяжется с вами в течение 15 минут прямо здесь в чате.")

	switch chat.BotState {
	case model.BotStateGreet, "":
		_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateMenu)
		return s.replyWithKeyboard(ctx, chat, s.greetingText(ctx, firstName), menuKeyboard())

	case model.BotStateMenu:
		// Пользователь печатает в меню — показываем меню ещё раз.
		return s.replyWithKeyboard(ctx, chat,
			"Выберите вариант на кнопке ниже 👇", menuKeyboard())

	case model.BotStateAskName:
		_ = s.chats.SetCollected(ctx, chat.ID, "collected_name", text)
		_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateAskProgram)
		return s.replyWithKeyboard(ctx, chat, askProgramText, programKeyboard(s.loadPrograms(ctx)))

	case model.BotStateAskProgram:
		// Если студент всё-таки набрал текст — пробуем распарсить как fallback,
		// иначе показываем кнопки снова.
		programs := s.loadPrograms(ctx)
		program := parseProgram(text, programs)
		if program == "" {
			return s.replyWithKeyboard(ctx, chat,
				"Пожалуйста, выберите программу кнопкой ниже 👇", programKeyboard(programs))
		}
		_ = s.chats.SetCollected(ctx, chat.ID, "collected_program", program)
		_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateAskPhone)
		return s.reply(ctx, chat, askPhoneText)

	case model.BotStateAskPhone:
		_ = s.chats.SetCollected(ctx, chat.ID, "collected_phone", text)
		// Создаём лида и привязываем его к чату
		nameParts := strings.SplitN(strings.TrimSpace(chat.CollectedName), " ", 2)
		fn := nameParts[0]
		ln := ""
		if len(nameParts) > 1 {
			ln = nameParts[1]
		}
		if fn == "" {
			fn = firstName
		}
		if fn == "" {
			fn = "Telegram-лид"
		}
		tgIDStr := fmt.Sprintf("%d", payload.Message.From.ID)

		req := &model.CreateLeadRequest{
			FirstName:  fn,
			LastName:   ln,
			Phone:      text,
			TelegramID: tgIDStr,
			UTMSource:  "telegram",
			UTMMedium:  "bot",
			UTMCampaign: "admissions_bot",
		}
		// Привязываем выбранную программу к лиду, если получится найти id.
		if s.programRepo != nil && chat.CollectedProgram != "" {
			if pid, ok, _ := s.programRepo.FindIDByName(ctx, chat.CollectedProgram); ok {
				req.ProgramID = &pid
			}
		}
		// Передаём также программу через email-поле было бы неверно;
		// программа сохранится в interactions как сводка ниже.
		lead, err := s.leadSvc.CreateLeadFromForm(ctx, req)
		if err != nil {
			return fmt.Errorf("create lead from telegram: %w", err)
		}
		_ = s.chats.SetLeadID(ctx, chat.ID, lead.ID)
		chat.LeadID = &lead.ID

		// Сводка анкеты в виде заметки
		summary := fmt.Sprintf("Telegram-бот собрал анкету:\n• ФИО: %s\n• Программа: %s\n• Телефон: %s",
			chat.CollectedName, chat.CollectedProgram, text)
		_, _ = s.intSvc.AddInteraction(ctx, &model.CreateInteractionRequest{
			LeadID:    &lead.ID,
			Type:      "note",
			Direction: "inbound",
			Content:   summary,
		})

		_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateManager)
		return s.reply(ctx, chat, handoffText)
	}

	return nil
}

// reply отправляет outbound сообщение и сохраняет его как interaction.
func (s *TelegramBotService) reply(ctx context.Context, chat *model.TelegramChat, text string) error {
	if s.client != nil {
		if err := s.client.SendMessage(ctx, chat.ChatID, text); err != nil {
			return err
		}
	}
	if chat.LeadID != nil {
		_, _ = s.intSvc.AddInteraction(ctx, &model.CreateInteractionRequest{
			LeadID:    chat.LeadID,
			Type:      "messenger",
			Direction: "outbound",
			Content:   text,
		})
	}
	return nil
}

// replyWithKeyboard — то же что reply, но с inline-клавиатурой.
func (s *TelegramBotService) replyWithKeyboard(ctx context.Context, chat *model.TelegramChat, text string, rows [][]telegram.InlineButton) error {
	if s.client != nil {
		if err := s.client.SendMessageWithKeyboard(ctx, chat.ChatID, text, rows); err != nil {
			return err
		}
	}
	if chat.LeadID != nil {
		_, _ = s.intSvc.AddInteraction(ctx, &model.CreateInteractionRequest{
			LeadID:    chat.LeadID,
			Type:      "messenger",
			Direction: "outbound",
			Content:   text,
		})
	}
	return nil
}

// handleCallback обрабатывает клик inline-кнопки.
// Сейчас поддерживается только программа: data="program:<index>".
func (s *TelegramBotService) handleCallback(ctx context.Context, cb *model.TelegramCallbackQuery) error {
	if cb.Message == nil || cb.Message.Chat.ID == 0 {
		return nil
	}
	chatID := int64(cb.Message.Chat.ID)
	chat, err := s.chats.GetByChatID(ctx, chatID)
	if err != nil {
		return err
	}
	if chat == nil {
		// Без существующего чата кнопка пришла бы только если /start ещё не нажимали — игнорируем.
		return nil
	}

	// Студент уже под управлением менеджера — кнопки игнорируем.
	if !chat.BotActive || chat.BotState == model.BotStateManager {
		if s.client != nil {
			_ = s.client.AnswerCallbackQuery(ctx, cb.ID, "")
		}
		return nil
	}

	ack := func(msg string) {
		if s.client != nil {
			_ = s.client.AnswerCallbackQuery(ctx, cb.ID, msg)
		}
	}

	// ── Главное меню (menu:*) ─────────────────────────────
	if strings.HasPrefix(cb.Data, "menu:") {
		choice := strings.TrimPrefix(cb.Data, "menu:")
		ack("")
		switch choice {
		case "admission":
			body := s.setting(ctx, "faq_admission",
				"Условия поступления уточнит менеджер.")
			_ = s.reply(ctx, chat, body)
			return s.replyWithKeyboard(ctx, chat, "Что ещё подсказать?", menuKeyboard())
		case "event":
			body := s.setting(ctx, "faq_event",
				"Расскажите, на какое мероприятие хотите попасть — менеджер ответит.")
			_ = s.reply(ctx, chat, body)
			return s.replyWithKeyboard(ctx, chat, "Что ещё подсказать?", menuKeyboard())
		case "eet":
			body := s.setting(ctx, "faq_eet",
				"EET (English Entry Test): расскажем расписание и запишем на ближайшую дату.")
			_ = s.reply(ctx, chat, body)
			return s.replyWithKeyboard(ctx, chat, "Что ещё подсказать?", menuKeyboard())
		case "apply":
			askName := s.setting(ctx, "ask_name", "Как вас зовут? Напишите, пожалуйста, ФИО.")
			_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateAskName)
			return s.reply(ctx, chat, askName)
		default:
			return nil
		}
	}

	// ── Выбор программы ───────────────────────────────────
	if chat.BotState != model.BotStateAskProgram {
		ack("")
		return nil
	}
	if !strings.HasPrefix(cb.Data, "program:") {
		ack("")
		return nil
	}
	programs := s.loadPrograms(ctx)
	idxStr := strings.TrimPrefix(cb.Data, "program:")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 || idx >= len(programs) {
		ack("Не распознан выбор")
		return nil
	}
	program := programs[idx]
	ack("")

	askPhoneText := s.setting(ctx, "ask_phone",
		"Отлично! Оставьте, пожалуйста, ваш номер телефона для связи (например, +992 90 123 45 67).")

	_ = s.chats.SetCollected(ctx, chat.ID, "collected_program", program)
	_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateAskPhone)
	_ = s.reply(ctx, chat, fmt.Sprintf("Выбрано: <b>%s</b>", program))
	return s.reply(ctx, chat, askPhoneText)
}

// ManagerSend — отправка сообщения вручную из CRM в Telegram.
// Автоматически переключает чат в состояние manager (перехват).
func (s *TelegramBotService) ManagerSend(ctx context.Context, leadID int, userID int, text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("text is empty")
	}
	chat, err := s.chats.GetByLeadID(ctx, leadID)
	if err != nil {
		return err
	}
	if chat == nil {
		return fmt.Errorf("no telegram chat for lead %d", leadID)
	}
	if s.client == nil {
		return fmt.Errorf("telegram bot token is not configured")
	}
	if err := s.client.SendMessage(ctx, chat.ChatID, text); err != nil {
		return err
	}
	// Помечаем чат как handoff
	_ = s.chats.UpdateState(ctx, chat.ID, model.BotStateManager)
	uid := userID
	_, _ = s.intSvc.AddInteraction(ctx, &model.CreateInteractionRequest{
		LeadID:    &leadID,
		Type:      "messenger",
		Direction: "outbound",
		Content:   text,
		CreatedBy: &uid,
	})
	s.publishMessage(&leadID, "outbound", text)
	return nil
}

// Takeover — менеджер вручную перехватывает чат (бот замолкает).
func (s *TelegramBotService) Takeover(ctx context.Context, leadID int) error {
	chat, err := s.chats.GetByLeadID(ctx, leadID)
	if err != nil {
		return err
	}
	if chat == nil {
		return fmt.Errorf("no telegram chat for lead %d", leadID)
	}
	return s.chats.UpdateState(ctx, chat.ID, model.BotStateManager)
}

// GetChatStatus возвращает текущее состояние чата для UI.
func (s *TelegramBotService) GetChatStatus(ctx context.Context, leadID int) (*model.TelegramChat, error) {
	return s.chats.GetByLeadID(ctx, leadID)
}
