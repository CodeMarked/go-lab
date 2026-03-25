import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { Observable } from 'rxjs';
import { AuthService } from './auth.service';
import { environment } from '../environments/environment';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent {
  title = 'Platform admin';
  loggedIn$: Observable<boolean>;
  useBootstrapAuth = environment.useBootstrapAuth;

  constructor(
    public auth: AuthService,
    private router: Router
  ) {
    this.loggedIn$ = this.auth.isLoggedIn();
  }

  logout(): void {
    this.auth.logout().subscribe(() => {
      void this.router.navigateByUrl('/login');
    });
  }
}
