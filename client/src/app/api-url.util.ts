import { environment } from '../environments/environment';

/**
 * Whether this HTTP request targets our API (apply credentials + CSRF).
 * Empty apiBaseUrl means same-origin paths (/api/v1/...) via nginx or dev-server proxy.
 */
export function isOurApiRequest(url: string): boolean {
  const base = (environment.apiBaseUrl || '').trim();
  if (base.length > 0) {
    return url.startsWith(base);
  }
  return url.startsWith('/api/') || url.includes('/api/v1/');
}
