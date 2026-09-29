const BASE = '/api';

export interface Recommendation {
  item_id: string;
  title: string;
  score: number;
  reason: string;
  labels: string[];
  category: string;
  label_similarity?: boolean;
}

export interface RecommendationResponse {
  user_id?: string;
  scenario: string;
  tenant_id: string;
  count: number;
  recommendations: Recommendation[];
  user_interests?: string[];
  interest_summary?: string;
  cold_start?: boolean;
}

export interface DashboardStats {
  tenant_id: string;
  totalUsers: number;
  totalItems: number;
  totalFeedback: number;
  topCategories: { name: string; category: string; count: number }[];
  recentEvents?: { feedback_type: string; user_id: string; item_id: string; timestamp: string }[];
}

export interface SimilarResponse {
  source_item_id: string;
  source_title: string;
  scenario: string;
  tenant_id: string;
  used_label_sim?: boolean;
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

export interface StrategyMetrics {
  strategy: string;
  users: number;
  precision_at_k: number;
  recall_at_k: number;
  hit_rate_at_k: number;
  mrr_at_k: number;
  ndcg_at_k: number;
}

export interface EvaluationResponse {
  tenant_id: string;
  category: string;
  k: number;
  evaluated_users: number;
  holdout: string;
  note: string;
  best_strategy: StrategyMetrics;
  metrics: StrategyMetrics[];
}

export interface EvaluatedRecommendationResponse extends RecommendationResponse {
  selected_strategy: string;
  selection_metric: string;
  evaluation: StrategyMetrics[];
}

export interface InterestOption {
  label: string;
  category: string;
  emoji: string;
}

export interface InterestOptionsResponse {
  interests: InterestOption[];
  min_pick: number;
  max_pick: number;
  tenant_id: string;
}

export interface OnboardResponse {
  user_id: string;
  interests: string[];
  tenant_id: string;
  cold_start: boolean;
  message: string;
  news: Recommendation[];
  products: Recommendation[];
  expo: Recommendation[];
}

function withRefresh(path: string): string {
  const separator = path.includes('?') ? '&' : '?';
  return `${path}${separator}refresh=${Date.now()}`;
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    cache: 'no-store',
    ...options,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${path} failed ${res.status}: ${text}`);
  }
  return res.json() as Promise<T>;
}

export const api = {
  // Onboarding (cold-start)
  getInterestOptions: () =>
    apiFetch<InterestOptionsResponse>('/onboard/interests'),

  postOnboard: (userId: string, interests: string[]) =>
    apiFetch<OnboardResponse>('/onboard', {
      method: 'POST',
      body: JSON.stringify({ userId, interests }),
    }),

  // Dashboard
  getDashboard: () => apiFetch<DashboardStats>('/demo/dashboard'),

  getEvaluation: (category = 'news', k = 10, users = 25) =>
    apiFetch<EvaluationResponse>(`/demo/evaluation?category=${category}&k=${k}&users=${users}`),

  // News
  getTrendingNews: (limit = 10) =>
    apiFetch<RecommendationResponse>(withRefresh(`/demo/trending-news?limit=${limit}`)),

  getPersonalizedNews: (userId: string, limit = 10) =>
    apiFetch<RecommendationResponse>(withRefresh(`/demo/personalized-news/${userId}?limit=${limit}`)),

  getEvaluatedNews: (userId: string, limit = 10, users = 25) =>
    apiFetch<EvaluatedRecommendationResponse>(withRefresh(`/demo/evaluated-news/${userId}?limit=${limit}&users=${users}`)),

  // Products
  getProductRecommendations: (userId: string, limit = 10) =>
    apiFetch<RecommendationResponse>(withRefresh(`/demo/products/user/${userId}/recommendations?limit=${limit}`)),

  getSimilarProducts: (itemId: string, limit = 10) =>
    apiFetch<SimilarResponse>(withRefresh(`/demo/products/item/${itemId}/similar?limit=${limit}`)),

  // Expo
  getExpoBooths: (buyerId: string, limit = 8) =>
    apiFetch<RecommendationResponse>(withRefresh(`/demo/expo/buyer/${buyerId}/booths?limit=${limit}`)),

  getExpoSuppliers: (buyerId: string, limit = 8) =>
    apiFetch<RecommendationResponse>(withRefresh(`/demo/expo/buyer/${buyerId}/suppliers?limit=${limit}`)),

  getExpoSessions: (buyerId: string, limit = 8) =>
    apiFetch<RecommendationResponse>(withRefresh(`/demo/expo/buyer/${buyerId}/sessions?limit=${limit}`)),

  getTrendingBooths: (limit = 10) =>
    apiFetch<RecommendationResponse>(withRefresh(`/demo/expo/trending?limit=${limit}`)),

  getSellerBuyers: (sellerId: string, limit = 8) =>
    apiFetch<BuyerResponse>(withRefresh(`/demo/expo/seller/${sellerId}/buyers?limit=${limit}`)),

  // Events
  postEvent: (feedbackType: string, userId: string, itemId: string) =>
    apiFetch<{ success: boolean; message: string }>('/events', {
      method: 'POST',
      body: JSON.stringify({ feedbackType, userId, itemId }),
    }),

  loadData: () =>
    apiFetch<Record<string, unknown>>('/demo/load-data', { method: 'POST', body: '{}' }),
};
