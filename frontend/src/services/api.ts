import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Function to set auth token (called from components with Clerk token)
export const setAuthToken = (token: string | null) => {
  if (token) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    console.log('[api] Auth token set:', `Bearer ${token.substring(0, 30)}...`);
  } else {
    delete api.defaults.headers.common['Authorization'];
    console.log('[api] Auth token cleared');
  }
};

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Don't redirect on 401 - Clerk handles auth
    // Just reject the promise so the caller can handle it
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
  create: (data: Record<string, unknown>) =>
    api.post('/vehicles', data),
  update: (id: number, data: Record<string, unknown>) =>
    api.put(`/vehicles/${id}`, data),
  delete: (id: number) =>
    api.delete(`/vehicles/${id}`),
  submit: (id: number) =>
    api.post(`/vehicles/${id}/submit`),
  getUploadUrl: (vehicleId: number, filename: string, contentType: string) =>
    api.post(`/vehicles/${vehicleId}/upload-url`, { filename, content_type: contentType }),
  addImage: (vehicleId: number, s3Key: string, url: string, isPrimary: boolean) =>
    api.post(`/vehicles/${vehicleId}/images`, { s3_key: s3Key, url, is_primary: isPrimary }),
};

export const auctionsApi = {
  list: (params?: Record<string, string>) =>
    api.get('/auctions', { params }),
  get: (id: number) => api.get(`/auctions/${id}`),
  placeBid: (auctionId: number, amount: number) =>
    api.post(`/auctions/${auctionId}/bid`, { amount }),
  getBids: (auctionId: number) =>
    api.get(`/auctions/${auctionId}/bids`),
  getActive: () => api.get('/auctions?status=active'),
  getEndingSoon: () => api.get('/auctions?ending_soon=true'),
};

export const watchlistApi = {
  list: () => api.get('/watchlist'),
  add: (auctionId: number) => api.post('/watchlist', { auction_id: auctionId }),
  remove: (auctionId: number) => api.delete(`/watchlist/${auctionId}`),
};

export const notificationsApi = {
  list: () => api.get('/notifications'),
  markRead: (id: number) => api.put(`/notifications/${id}/read`),
  markAllRead: () => api.put('/notifications/read-all'),
  unreadCount: () => api.get('/notifications/unread-count'),
};

export const ordersApi = {
  list: () => api.get('/orders'),
  get: (id: number) => api.get(`/orders/${id}`),
};
