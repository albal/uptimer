import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Login } from './pages/Login';

describe('Login Page', () => {
  it('renders login buttons', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );

    expect(screen.getByText(/Sign in with Google/i)).toBeDefined();
    expect(screen.getByText(/Sign in with Microsoft/i)).toBeDefined();
    expect(screen.getByText(/Sign in with Apple/i)).toBeDefined();
  });

  it('shows the branding', () => {
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>
    );

    expect(screen.getByText(/Uptimer/i)).toBeDefined();
  });
});
