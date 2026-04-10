import React from 'react';
import { Outlet } from 'react-router-dom';
import { Sidebar } from './Sidebar';
import { Header } from './Header';

export function Layout() {
  return (
    <div className="app-layout">
      <Sidebar />
      <Header />
      <main className="app-main">
        <Outlet />
      </main>
    </div>
  );
}
