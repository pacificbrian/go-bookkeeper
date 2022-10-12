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
  static targets = [ "cashflowTableRow", "cashflowAmount", "cashflowNewAmount" ]

  cashflowDelete$ = new Subject();
  cashflowPut$ = new Subject();

  connect() {
    console.log("Stimulus connected!", this.element);

    this.cashflowDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((cashflowID) => {
          console.log("RXJS:ajax:DELETE: ", [cashflowID])
          return ajax({
            method: 'DELETE',
            url: '/cash_flows/'+cashflowID,
            responseType: 'json',
          });
        }),
        map((response) => {
          return response.response;
        })
      )
      .subscribe((response) => {
        console.log(response)
      })

    this.cashflowPut$
      .pipe(
        distinctUntilChanged(),
        switchMap((cashflowSet) => {
          console.log("RXJS:ajax:PUT: ", cashflowSet)
          return ajax({
            method: 'PUT',
            url: '/cash_flows/'+cashflowSet[0],
            responseType: 'json',
            body: {
              key: cashflowSet[1],
              value: +cashflowSet[2]
            }
          });
        }),
        map((response) => {
          return response.response;
        })
      )
      .subscribe((response) => {
        console.log(response)
        // for GETs:
        //this.someDivClass.innerHTML = response;
      })
  }

  disconnect() {
    this.cashflowDelete$.unsubscribe();
    this.cashflowPut$.unsubscribe();
  }

  actionDelete(event) {
    let target = event.currentTarget
    let cashflowID = target.getAttribute('data-cashflow-id')
    console.log("Stimulus: actionDelete", cashflowID)
    event.preventDefault()
    // add to RXJS stream processed with cashflowDelete.pipe above
    this.cashflowDelete$.next(cashflowID)

    // hide deleted CashFlow row in table
    const cashflowRows = this.cashflowTableRowTargets;
    for (let r in cashflowRows) {
      let id = cashflowRows[r].getAttribute('data-cashflow-id')
      if (id == cashflowID) {
        cashflowRows[r].hidden = 1
      }
    }
  }

  // using click event
  actionEditAmount(event) {
    // hide value in table
    let displayAmount = event.currentTarget
    let cashflowID = displayAmount.getAttribute('data-cashflow-id')
    displayAmount.hidden = 1

    console.log("Stimulus: actionEditAmount", cashflowID)

    // unhide input field so user can change value
    const inputAmounts = this.cashflowNewAmountTargets
    for (let i in inputAmounts) {
      let id = inputAmounts[i].getAttribute('data-cashflow-id')
      if (id == cashflowID) {
        inputAmounts[i].hidden = 0
        inputAmounts[i].autofocus = 1
      }
    }
  }

  // using keydown event (future: cancel with unclick?)
  actionPutNewAmount(event) {
    let send_amount = 0
    if (event.keyCode == 13) { // ENTER
      send_amount = 1
    } else if (event.keyCode == 27) { // ESCAPE
      send_amount = 0
    } else {
      return
    }
    console.log("Stimulus: keyCode is ", event.keyCode)

    // get new value and hide input field
    let inputAmount = event.currentTarget
    let cashflowID = inputAmount.getAttribute('data-cashflow-id')
    let newAmount = inputAmount.value
    inputAmount.hidden = 1

    if (send_amount) {
      console.log("Stimulus: actionPutNewAmount", cashflowID, newAmount)
    }

    // unhide and update value in table (if changed)
    const displayAmounts = this.cashflowAmountTargets;
    for (let i in displayAmounts) {
      let id = displayAmounts[i].getAttribute('data-cashflow-id')
      if (id == cashflowID) {
        if (send_amount) {
          displayAmounts[i].innerHTML = "$"+(parseFloat(newAmount).toFixed(2))
        }
        displayAmounts[i].hidden = 0
      }
    }

    // add to RXJS stream processed with cashflowPut.pipe
    this.cashflowPut$.next([cashflowID, "amount", newAmount]);
  }
}
