import { useContext } from "react";

import { AuthContext } from "@/lib/auth/auth-context";

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};
