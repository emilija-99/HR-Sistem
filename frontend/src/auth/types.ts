export type User = {
  id: number;
  email: string;
  role: string;
};

export type AuthContextType = {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  login: (data: { token: string; user: User }) => void;
  logout: () => void;
};
