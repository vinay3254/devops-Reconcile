import React, { createContext, useContext, useState, useEffect } from 'react';
import { authApi } from '../api/client';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [tenant, setTenant] = useState(null);
  const [token, setToken] = useState(localStorage.getItem('statuspage_token'));
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const savedUser = localStorage.getItem('statuspage_user');
    const savedTenant = localStorage.getItem('statuspage_tenant');
    if (savedUser && savedTenant && token) {
      setUser(JSON.parse(savedUser));
      setTenant(JSON.parse(savedTenant));
    }
    setLoading(false);
  }, [token]);

  const handleAuthSuccess = (res) => {
    localStorage.setItem('statuspage_token', res.token);
    localStorage.setItem('statuspage_user', JSON.stringify(res.user));
    localStorage.setItem('statuspage_tenant', JSON.stringify(res.tenant));
    setToken(res.token);
    setUser(res.user);
    setTenant(res.tenant);
  };

  const login = async (email, password) => {
    const res = await authApi.login(email, password);
    handleAuthSuccess(res);
  };

  const signup = async (tenantName, email, password) => {
    const res = await authApi.signup(tenantName, email, password);
    handleAuthSuccess(res);
  };

  const logout = () => {
    localStorage.removeItem('statuspage_token');
    localStorage.removeItem('statuspage_user');
    localStorage.removeItem('statuspage_tenant');
    setToken(null);
    setUser(null);
    setTenant(null);
  };

  return (
    <AuthContext.Provider value={{ user, tenant, token, login, signup, logout, loading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
