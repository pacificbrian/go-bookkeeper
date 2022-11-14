/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { Controller } from '@hotwired/stimulus';
import { Subject } from 'rxjs';
import { ajax } from 'rxjs/ajax';
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators';

export default class extends Controller {
  tradeDelete$ = new Subject();

  connect() {
    console.log("Stimulus[TRADE] connected!", this.element);

    this.tradeDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((tradeID) => {
          console.log("RXJS[TRADE]:ajax:DELETE: ", [tradeID])
          return ajax({
            method: 'DELETE',
            url: '/trades/'+tradeID,
            responseType: 'json'
          });
        }),
        map((response) => {
          return response.response;
        })
      )
      .subscribe((response) => {
        console.log(response)
      })
  }

  disconnect() {
    this.tradeDelete$.unsubscribe();
  }

  actionDelete(event) {
    let target = event.currentTarget
    let tradeID = target.getAttribute('data-trade-id')
    console.log("Stimulus[TRADE]: actionDelete", tradeID)
    event.preventDefault()
    // add to RXJS stream processed with tradeDelete.pipe above
    this.tradeDelete$.next(tradeID)
  }
}
