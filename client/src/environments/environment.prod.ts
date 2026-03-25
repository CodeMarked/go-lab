export const environment = {
  production: true,
  /** Empty = same-origin; Docker nginx proxies `/api` and `/healthz` / `/readyz` to the backend. */
  apiBaseUrl: '',
  useBootstrapAuth: false,
  csrfCookieName: 'gl_csrf',
  csrfHeaderName: 'X-CSRF-Token',
  sessionRefreshIntervalMs: 14 * 60 * 1000
};
