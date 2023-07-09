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
  static targets = [ "taxEntryRow", "taxReturnRow" ]
  taxEntryDelete$ = new Subject();
  taxReturnDelete$ = new Subject();
  taxPut$ = new Subject();

  connect() {
    console.log("Stimulus[TAX] connected!", this.element);

    this.taxEntryDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((taxID) => {
          console.log("RXJS[TAX ENTRY]:ajax:DELETE: ", [taxID])
          return ajax({
            method: 'DELETE',
            url: '/tax_entries/'+taxID,
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

    this.taxReturnDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((taxID) => {
          console.log("RXJS[TAX RETURN]:ajax:DELETE: ", [taxID])
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
    this.taxEntryDelete$.unsubscribe();
    this.taxReturnDelete$.unsubscribe();
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

    if (!confirm("Are you sure?"))
      return
    // add to RXJS stream processed with taxReturnDelete.pipe above
    this.taxReturnDelete$.next(taxID)

    // hide deleted TaxReturn row in table
    const taxReturnRows = this.taxReturnRowTargets;
    for (let i in taxReturnRows) {
      let id = taxReturnRows[i].getAttribute('data-tax-id')
      if (id == taxID) {
        taxReturnRows[i].hidden = 1
      }
    }
  }

  actionEntryDelete(event) {
    let target = event.currentTarget
    let taxID = target.getAttribute('data-tax-entry-id')
    console.log("Stimulus[TAX]: actionEntryDelete", taxID)
    event.preventDefault()

    if (!confirm("Are you sure?"))
      return
    // add to RXJS stream processed with taxEntryDelete.pipe above
    this.taxEntryDelete$.next(taxID)

    // hide deleted TaxEntry row in table
    const taxEntryRows = this.taxEntryRowTargets;
    for (let i in taxEntryRows) {
      let id = taxEntryRows[i].getAttribute('data-tax-entry-id')
      if (id == taxID) {
        taxEntryRows[i].hidden = 1
      }
    }
  }
}
