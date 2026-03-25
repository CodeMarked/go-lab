// This file can be replaced during build by using the `fileReplacements` array.
// The list of file replacements can be found in `angular.json`.

export const environment = {
  production: false,
  /**
   * Empty = same-origin `/api/...` (ng serve uses proxy.conf.json → :5000).
   * Set `http://localhost:5000` only if you accept that cookie sessions will not work (cross-site Lax).
   */
  apiBaseUrl: '',
  /** When true, app init calls POST /auth/bootstrap (dev bridge). When false, cookie session + login/register. */
  useBootstrapAuth: false,
  csrfCookieName: 'gl_csrf',
  csrfHeaderName: 'X-CSRF-Token',
  /**
   * Cookie-session mode only: interval for POST /auth/refresh. Keep below API
   * SESSION_IDLE_TTL_SECONDS (default 1800s) so the SPA slides idle before expiry.
   */
  sessionRefreshIntervalMs: 14 * 60 * 1000
};
