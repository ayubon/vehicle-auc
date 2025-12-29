import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;

// API functions
export const authApi = {
  login: (email: string, password: string) =>
    api.post('/auth/login', { email, password }),
  register: (data: { email: string; password: string; first_name?: string; last_name?: string }) =>
    api.post('/auth/register', data),
  logout: () => api.post('/auth/logout'),
  me: () => api.get('/auth/me'),
};

export const vehiclesApi = {
  list: (params?: Record<string, string>) =>
    api.get('/vehicles', { params }),
  get: (id: number) => api.get(`/vehicles/${id}`),
  create: (data: FormData) =>
    api.post('/vehicles', data, { headers: { 'Content-Type': 'multipart/form-data' } }),
};

export const auctionsApi = {
  list: (params?: Record<string, string>) =>
    api.get('/auctions', { params }),
  get: (id: number) => api.get(`/auctions/${id}`),
  placeBid: (auctionId: number, amount: number) =>
    api.post(`/auctions/${auctionId}/bid`, { amount }),
  getBids: (auctionId: number) =>
    api.get(`/auctions/${auctionId}/bids`),
};

export const ordersApi = {
  list: () => api.get('/orders'),
  get: (id: number) => api.get(`/orders/${id}`),
};
