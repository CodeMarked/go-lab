import { Component, OnInit } from '@angular/core';
import { User } from '../user';
import { UserService } from '../user.service';
import { environment } from '../../environments/environment';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class DashboardComponent implements OnInit {
  users: User[] = [];
  /** For health/ready links: empty apiBaseUrl → same-origin paths. */
  readonly apiBaseUrl = environment.apiBaseUrl;
  readonly apiBaseLabel =
    environment.apiBaseUrl.trim() !== ''
      ? environment.apiBaseUrl
      : '(same origin)';

  constructor(private userService: UserService) {}

  ngOnInit(): void {
    this.getUsers();
  }

  getUsers(): void {
    this.userService.getUsers().subscribe((users) =>
      (this.users = users
        .sort((a, b) => b.pennies - a.pennies)
        .slice(0, 3))
    );
  }
}
