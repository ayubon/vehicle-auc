import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

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
};

export const ordersApi = {
  list: () => api.get('/orders'),
  get: (id: number) => api.get(`/orders/${id}`),
};
