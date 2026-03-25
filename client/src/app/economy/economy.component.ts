import { Component, OnInit } from '@angular/core';
import { PlatformService } from '../platform.service';
import { EconomyLedgerListData } from './economy.models';

@Component({
  selector: 'app-economy',
  templateUrl: './economy.component.html',
  styleUrls: ['./economy.component.css']
})
export class EconomyComponent implements OnInit {
  data: EconomyLedgerListData | null = null;

  limit = 50;
  beforeId = ''; // keyset pagination: rows with id < before_id (newest first)
  platformUserId = '';
  eventType = '';
  from = '';
  to = '';

  constructor(private platform: PlatformService) {}

  ngOnInit(): void {
    this.load();
  }

  apply(): void {
    this.load();
  }

  private load(): void {
    const uid = this.platformUserId.trim();
    let lim = Number(this.limit);
    if (Number.isNaN(lim) || lim < 1) {
      lim = 50;
    }
    if (lim > 100) {
      lim = 100;
    }
    const params: {
      limit: number;
      before_id?: number;
      platform_user_id?: number;
      event_type?: string;
      from?: string;
      to?: string;
    } = { limit: lim };
    const bid = this.beforeId.trim();
    if (bid !== '') {
      const n = parseInt(bid, 10);
      if (!Number.isNaN(n) && n > 0) {
        params.before_id = n;
      }
    }
    if (uid !== '') {
      const n = parseInt(uid, 10);
      if (!Number.isNaN(n) && n > 0) {
        params.platform_user_id = n;
      }
    }
    const et = this.eventType.trim();
    if (et !== '') {
      params.event_type = et;
    }
    if (this.from.trim() !== '') {
      params.from = this.from.trim();
    }
    if (this.to.trim() !== '') {
      params.to = this.to.trim();
    }
    this.platform.getEconomyLedger(params).subscribe((d) => (this.data = d));
  }
}
