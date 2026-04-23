import { Outlet } from 'react-router-dom';

import { Sidebar } from './Sidebar';
import { Topbar } from './Topbar';

export function AppShell() {
  return (
    <div className="app-shell">
      <Sidebar />
      <div className="main-panel">
        <Topbar />
        <main className="content">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
