const BASE = '/api';

export interface Recommendation {
  item_id: string;
  title: string;
  score: number;
  reason: string;
  labels: string[];
  category: string;
}

export interface RecommendationResponse {
  user_id?: string;
  scenario: string;
  tenant_id: string;
  count: number;
  recommendations: Recommendation[];
  user_interests?: string[];
  interest_summary?: string;
}

export interface DashboardStats {
  tenant_id: string;
  totalUsers: number;
  totalItems: number;
  totalFeedback: number;
  topCategories: { name: string; category: string; count: number }[];
}

export interface SimilarResponse {
  source_item_id: string;
  source_title: string;
  scenario: string;
  tenant_id: string;
  count: number;
  recommendations: Recommendation[];
}

export interface Buyer {
  user_id: string;
  name: string;
  score: number;
  reason: string;
  labels: string[];
}

export interface BuyerResponse {
  seller_id: string;
  seller_name?: string;
  seller_labels?: string[];
  scenario: string;
  tenant_id: string;
  count: number;
  buyers: Buyer[];
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${path} failed ${res.status}: ${text}`);
  }
  return res.json() as Promise<T>;
}

export const api = {
  getDashboard: () => apiFetch<DashboardStats>('/demo/dashboard'),

  getTrendingNews: (limit = 10) =>
    apiFetch<RecommendationResponse>(`/demo/trending-news?limit=${limit}`),

  getPersonalizedNews: (userId: string, limit = 10) =>
    apiFetch<RecommendationResponse>(`/demo/personalized-news/${userId}?limit=${limit}`),

  getProductRecommendations: (userId: string, limit = 10) =>
    apiFetch<RecommendationResponse>(`/demo/products/user/${userId}/recommendations?limit=${limit}`),

  getSimilarProducts: (itemId: string, limit = 10) =>
    apiFetch<SimilarResponse>(`/demo/products/item/${itemId}/similar?limit=${limit}`),

  getExpoBooths: (buyerId: string, limit = 8) =>
    apiFetch<RecommendationResponse>(`/demo/expo/buyer/${buyerId}/booths?limit=${limit}`),

  getExpoSuppliers: (buyerId: string, limit = 8) =>
    apiFetch<RecommendationResponse>(`/demo/expo/buyer/${buyerId}/suppliers?limit=${limit}`),

  getExpoSessions: (buyerId: string, limit = 8) =>
    apiFetch<RecommendationResponse>(`/demo/expo/buyer/${buyerId}/sessions?limit=${limit}`),

  getTrendingBooths: (limit = 10) =>
    apiFetch<RecommendationResponse>(`/demo/expo/trending?limit=${limit}`),

  getSellerBuyers: (sellerId: string, limit = 8) =>
    apiFetch<BuyerResponse>(`/demo/expo/seller/${sellerId}/buyers?limit=${limit}`),

  postEvent: (feedbackType: string, userId: string, itemId: string) =>
    apiFetch<{ success: boolean; message: string }>('/events', {
      method: 'POST',
      body: JSON.stringify({ feedbackType, userId, itemId }),
    }),

  loadData: () =>
    apiFetch<Record<string, unknown>>('/demo/load-data', { method: 'POST', body: '{}' }),
};
