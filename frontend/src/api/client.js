const API_BASE_URL = 'http://localhost:8080/api/v1';

export async function apiRequest(endpoint, method = 'GET', body = null) {
  const token = localStorage.getItem('statuspage_token');
  const headers = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const options = {
    method,
    headers,
  };

  if (body) {
    options.body = JSON.stringify(body);
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, options);
  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || 'API Request Failed');
  }

  return data;
}

export const authApi = {
  signup: (tenantName, email, password) =>
    apiRequest('/auth/signup', 'POST', { tenant_name: tenantName, email, password }),
  login: (email, password) =>
    apiRequest('/auth/login', 'POST', { email, password }),
};

export const monitorApi = {
  list: () => apiRequest('/monitors', 'GET'),
  get: (id) => apiRequest(`/monitors/${id}`, 'GET'),
  create: (data) => apiRequest('/monitors', 'POST', data),
  delete: (id) => apiRequest(`/monitors/${id}`, 'DELETE'),
  getPublicStatus: (tenantId) => apiRequest(`/public/status/${tenantId}`, 'GET'),
};
