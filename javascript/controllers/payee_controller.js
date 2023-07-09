/*
 * SPDX-FileCopyrightText: 2023 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { Controller } from '@hotwired/stimulus';
import { Subject } from 'rxjs';
import { ajax } from 'rxjs/ajax';
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators';

export default class extends Controller {
  payeeDelete$ = new Subject();

  connect() {
    console.log("Stimulus[PAYEE] connected!", this.element);

    this.payeeDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((payeeID) => {
          console.log("RXJS[PAYEE]:ajax:DELETE: ", [payeeID])
          return ajax({
            method: 'DELETE',
            url: '/payees/'+payeeID,
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
    this.payeeDelete$.unsubscribe();
  }

  actionDelete(event) {
    let target = event.currentTarget
    let payeeID = target.getAttribute('data-payee-id')
    console.log("Stimulus[PAYEE]: actionDelete", payeeID)
    event.preventDefault()

    if (!confirm("Are you sure?"))
      return
    // add to RXJS stream processed with payeeDelete.pipe above
    this.payeeDelete$.next(payeeID)
  }
}
