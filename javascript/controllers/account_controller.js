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
  accountDelete$ = new Subject();

  connect() {
    console.log("Stimulus[ACCOUNT] connected!", this.element);

    this.accountDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((accountID) => {
          console.log("RXJS[ACCOUNT]:ajax:DELETE: ", [accountID])
          return ajax({
            method: 'DELETE',
            url: '/accounts/'+accountID,
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
    this.accountDelete$.unsubscribe();
  }

  actionDelete(event) {
    let target = event.currentTarget
    let accountID = target.getAttribute('data-account-id')
    console.log("Stimulus[ACCOUNT]: actionDelete", accountID)
    event.preventDefault()

    if (!confirm("Are you sure?"))
      return
    // add to RXJS stream processed with accountDelete.pipe above
    this.accountDelete$.next(accountID)
  }
}
