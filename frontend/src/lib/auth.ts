import { jwtDecode } from 'jwt-decode';

export interface User {
  id: number;
  email: string;
}

/**
 * Set the JWT token in a cookie
 */
export function setToken(token: string): void {
  // Store token in a cookie that expires in 24 hours
  document.cookie = `token=${token}; path=/; max-age=${60 * 60 * 24}; SameSite=Lax`;
}

/**
 * Get the JWT token from cookies
 */
export function getToken(): string | null {
  const match = document.cookie.match(/(^| )token=([^;]+)/);
  return match ? match[2] : null;
}

/**
 * Remove the JWT token from cookies
 */
export function removeToken(): void {
  document.cookie = 'token=; path=/; max-age=0';
}

/**
 * Decode the JWT token to get user information
 * Note: This does not validate the token signature, only decodes the payload.
 * For security-critical operations, always validate the token on the backend.
 */
export function decodeToken(token: string): User | null {
  try {
    const decoded = jwtDecode<{ user_id: number; email: string }>(token);
    return {
      id: decoded.user_id,
      email: decoded.email
    };
  } catch {
    // FIXED: Removed the unused 'err' variable
    return null;
  }
}

/**
 * Check if the user is authenticated (has a valid token)
 */
export function isAuthenticated(): boolean {
  const token = getToken();
  if (!token) return false;

  // Check if token is expired by decoding and checking exp
  try {
    // FIXED: Passed the exact expected type instead of using 'any' or 'unknown'
    const decoded = jwtDecode<{ exp: number }>(token);
    return decoded.exp * 1000 > Date.now();
  } catch {
    return false;
  }
}