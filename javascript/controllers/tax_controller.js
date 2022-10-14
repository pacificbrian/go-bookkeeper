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
  taxDelete$ = new Subject();
  taxPut$ = new Subject();

  connect() {
    console.log("Stimulus[TAX] connected!", this.element);

    this.taxDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((taxID) => {
          console.log("RXJS[TAX]:ajax:DELETE: ", [taxID])
          return ajax({
            method: 'DELETE',
            url: '/taxes/'+taxID,
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

    this.taxPut$
      .pipe(
        distinctUntilChanged(),
        switchMap((taxID) => {
          console.log("RXJS[TAX]:ajax:PUT: ", [taxID])
          return ajax({
            method: 'PUT',
            url: '/taxes/'+taxID,
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
    this.taxDelete$.unsubscribe();
    this.taxPut$.unsubscribe();
  }

  actionCalculate(event) {
    let target = event.currentTarget
    let taxID = target.getAttribute('data-tax-id')
    console.log("Stimulus[TAX]: actionCalculate", taxID)
    event.preventDefault()
    // add to RXJS stream processed with taxPut.pipe above
    this.taxPut$.next(taxID)
  }

  actionDelete(event) {
    let target = event.currentTarget
    let taxID = target.getAttribute('data-tax-id')
    console.log("Stimulus[TAX]: actionDelete", taxID)
    event.preventDefault()
    // add to RXJS stream processed with taxDelete.pipe above
    this.taxDelete$.next(taxID)
  }
}
