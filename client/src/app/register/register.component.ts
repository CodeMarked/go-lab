import { Component } from '@angular/core';
import { Router } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { AuthService } from '../auth.service';

@Component({
  selector: 'app-register',
  templateUrl: './register.component.html',
  styleUrls: ['./register.component.css']
})
export class RegisterComponent {
  email = '';
  password = '';
  name = '';
  errorMsg = '';
  loading = false;

  constructor(
    private auth: AuthService,
    private router: Router
  ) {}

  submit(): void {
    this.errorMsg = '';
    this.loading = true;
    this.auth
      .register(this.email.trim(), this.password, this.name.trim())
      .subscribe({
        next: () => {
          this.loading = false;
          void this.router.navigateByUrl('/login');
        },
        error: (err: unknown) => {
          this.loading = false;
          if (err instanceof HttpErrorResponse) {
            const e = err.error?.error;
            this.errorMsg =
              e && typeof e.message === 'string'
                ? e.message
                : 'Registration failed';
          } else {
            this.errorMsg = 'Registration failed';
          }
        }
      });
  }
}
