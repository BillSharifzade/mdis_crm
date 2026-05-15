import { useState } from 'react';
import { Menu, Search, Bell } from 'lucide-react';

export default function Topbar({ breadcrumb, onMenuToggle, onNotifToggle, onSearch, searchValue = '', unreadCount = 0, hideSearch = false }) {
    const [searchVal, setSearchVal] = useState(searchValue);

    const handleSearch = (e) => {
        setSearchVal(e.target.value);
        onSearch(e.target.value);
    };

    return (
        <header className="topbar">
            <div className="topbar-left">
                <button className="topbar-menu-btn" onClick={onMenuToggle}>
                    <Menu size={18} />
                </button>
                <div className="breadcrumb">{breadcrumb}</div>
            </div>
            <div className="topbar-right">
                {!hideSearch && (
                    <div className="search-box">
                        <Search size={15} />
                        <input
                            type="text"
                            placeholder="Поиск лидов по имени или телефону..."
                            value={searchVal}
                            onChange={handleSearch}
                        />
                    </div>
                )}
                <button className="topbar-btn" onClick={onNotifToggle}>
                    <Bell size={16} />
                    {unreadCount > 0 && <span className="notif-dot"></span>}
                </button>
            </div>
        </header>
    );
}
