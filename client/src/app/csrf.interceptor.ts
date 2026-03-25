import { Injectable } from '@angular/core';
import {
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../environments/environment';
import { readCookie } from './csrf.util';
import { isOurApiRequest } from './api-url.util';

const CSRF_EXEMPT_SUBSTRINGS = [
  '/api/v1/auth/register',
  '/api/v1/auth/login',
  '/api/v1/auth/token',
  '/api/v1/auth/bootstrap'
];

@Injectable()
export class CsrfInterceptor implements HttpInterceptor {
  intercept(
    req: HttpRequest<unknown>,
    next: HttpHandler
  ): Observable<HttpEvent<unknown>> {
    if (!isOurApiRequest(req.url)) {
      return next.handle(req);
    }
    const m = req.method.toUpperCase();
    if (m === 'GET' || m === 'HEAD' || m === 'OPTIONS') {
      return next.handle(req);
    }
    if (CSRF_EXEMPT_SUBSTRINGS.some((s) => req.url.includes(s))) {
      return next.handle(req);
    }
    if (req.headers.has('Authorization')) {
      return next.handle(req);
    }
    const token = readCookie(environment.csrfCookieName);
    if (!token) {
      return next.handle(req);
    }
    return next.handle(
      req.clone({
        setHeaders: { [environment.csrfHeaderName]: token }
      })
    );
  }
}
