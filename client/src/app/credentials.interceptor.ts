import { Injectable } from '@angular/core';
import {
  HttpEvent,
  HttpHandler,
  HttpInterceptor,
  HttpRequest
} from '@angular/common/http';
import { Observable } from 'rxjs';
import { isOurApiRequest } from './api-url.util';

/** Send cookies on API calls (session + CSRF cookies). */
@Injectable()
export class CredentialsInterceptor implements HttpInterceptor {
  intercept(
    req: HttpRequest<unknown>,
    next: HttpHandler
  ): Observable<HttpEvent<unknown>> {
    if (isOurApiRequest(req.url)) {
      return next.handle(req.clone({ withCredentials: true }));
    }
    return next.handle(req);
  }
}
