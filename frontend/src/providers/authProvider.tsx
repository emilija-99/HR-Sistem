import axios from "axios";
import { createContext, useContext, useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";

type AuthContextType = {
  token: string | null;
  setToken: (token: string | null) => void;
};

/*
  create a context object that will be used
  to share the authentication state
  and functions between components
*/
const AuthContext = createContext<AuthContextType | undefined>(undefined);

type AuthProviderProps = {
  children: ReactNode;
};

/*
  serves as the provider for the authentication context
  it receives children as a prop, which represents the child
  component that will have access to the authentication context
*/
const AuthProvider = ({ children }: AuthProviderProps) => {
  /*
  token represents the authentication token
  from localStorage retrives the token value if it exists
*/
  const [token, setToken_] = useState<string | null>(
    localStorage.getItem("token"),
  );

  /*
    set the new token value
    it updates the token state using setToken_() and stores the token value
    in the local storage using localStorage.setItem()
  */
  const setToken = (newToken: string | null) => {
    setToken_(newToken);
  };

  /*
    useEffect() to set the default authorization header in axios and stores the token value
    in the local storage using localStorage.setItem()
  */
  useEffect(() => {
    if (token) {
      axios.defaults.headers.common["Authorization"] = "Bearer " + token;
      localStorage.setItem("token", token);
    } else {
      delete axios.defaults.headers.common["Authorization"];
      localStorage.removeItem("token");
    }
  }, [token]);

  /*
    create the memorized context using useMemo();
    - the context value includes the token adn setToken function
    - the token value is used as a dependency for memorization
  */
  const contextValue = useMemo(() => ({ token, setToken }), [token]);

  /*
    provide the authentication context to the child components
    - wrap the children component with the AuthContext.Provider,
    - pass the contextValue as the value prop of the provider
  */
  return (
    <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
  );
};

/*
  export useAuth hook for accessing the authentication context
  - useAuth is a custom hook that can be used in components to access the authentication context
*/
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
};

export default AuthProvider;
