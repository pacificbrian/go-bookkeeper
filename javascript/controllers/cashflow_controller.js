/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

// For DOM manipulation API reference, see:
//   https://developer.mozilla.org/en-US/docs/Web/API/HTML_DOM_API

import { Controller } from '@hotwired/stimulus';
import { Subject } from 'rxjs';
import { ajax } from 'rxjs/ajax';
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators';

export default class extends Controller {
  static targets = [ "cashflowTableRow", "cashflowTableRowBalance",
                     "cashflowAmount", "cashflowNewAmount" ]

  cashflowDelete$ = new Subject();
  cashflowPut$ = new Subject();

  connect() {
    console.log("Stimulus[CASHFLOW] connected!", this.element);

    this.cashflowDelete$
      .pipe(
        distinctUntilChanged(),
        switchMap((cashflowID) => {
          console.log("RXJS[CASHFLOW]:ajax:DELETE: ", [cashflowID])
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
          console.log("RXJS[CASHFLOW]:ajax:PUT: ", cashflowSet)
          return ajax({
            method: 'PUT',
            url: '/cash_flows/'+cashflowSet[0],
            responseType: 'json',
            body: {
              // server reads using KeyValue struct with string types
              key: cashflowSet[1],
              value: cashflowSet[2]
            }
          });
        }),
        map((response) => {
          return response.response;
        })
      )
      .subscribe((response) => {
        console.log("RXJS[CASHFLOW]:ajax:PUT reply: " + response)
        // for GETs:
        //this.someDivClass.innerHTML = response;
      })
  }

  disconnect() {
    this.cashflowDelete$.unsubscribe();
    this.cashflowPut$.unsubscribe();
  }

  actionApply(event) {
    let target = event.currentTarget
    let cashflowID = target.getAttribute('data-cashflow-id')
    console.log("Stimulus[CASHFLOW]: actionApply", cashflowID)
    event.preventDefault()
    // add to RXJS stream processed with cashflowPut.pipe above
    this.cashflowPut$.next([cashflowID, "apply", "1"]);
  }

  adjustBalances(tableIdx, adjustAmount) {
    // iterate and fixup Balance amounts until we pass modified row
    const cashflowBalances = this.cashflowTableRowBalanceTargets;
    for (let i in cashflowBalances) {
      if (parseInt(i) > tableIdx) {
        break
      }
      let oldAmountHTML = cashflowBalances[i].innerHTML
      let amountStart = oldAmountHTML.search(/[$]/)
      let oldBalance = parseFloat(oldAmountHTML.slice(amountStart+1))
      let newBalance = oldBalance + adjustAmount
      // overwrite c.Balance
      cashflowBalances[i].innerHTML = "$"+(newBalance.toFixed(2))
    }
    console.log("Stimulus[CASHFLOW]: UPDATE BALANCES[%d] adjustAmount: %f", tableIdx, adjustAmount)
  }

  // dynamically add row to CashFlow Ledger
  // options look to be directly via DOM API:
  //   https://developer.mozilla.org/en-US/docs/Web/API/HTMLTableElement/insertRow
  //   Prototyped below...
  // or more elegantly can do with turbo:
  //   https://turbo.hotwired.dev/handbook/introduction
  actionCreate(event) {
    let target = event.currentTarget

    //can block browser from sending POST and send CashFlow using RXJS pipe
    //  event.preventDefault()
    //or don't do above, and just have browser send POST and server return c.NoContent

    console.log("Stimulus[CASHFLOW]: actionCreate")

    // dynamically add row to CashFlow Ledger
    //
    //newCashFlow.Date = target.getAttribute('data-cashflow-date')
    //newCashFlow.PayeeName = target.getAttribute('data-cashflow-payee-name')
    //newCashFlow.Amount = target.getAttribute('data-cashflow-amount')
    //newCashFlow.Memo = target.getAttribute('data-cashflow-memo')
    // or see if HTML DOM API for reading form fields
    //
    // can I insert a row in existing table, sorted?
    // or destroy 'tbody' and rewrite it, plus regen Balances?
    //const cashflowRows = this.cashflowTableRowTargets;
    //for (let r in cashflowRows) {
    //  let date = cashflowRows[r].getAttribute('data-cashflow-date')
    //  if date > newCashFlow.Date
    //    break
    //}
    //
    //row = controllerTarget.insertRow(r)
    //for each newCashFlow field:
    //  cell = row.insertCell()
    //  cell.innerHTML = ""
  }

  actionDelete(event) {
    let target = event.currentTarget
    let cashflowID = target.getAttribute('data-cashflow-id')
    let adjustAmount = 0
    let tableIdx = -1

    console.log("Stimulus[CASHFLOW]: actionDelete", cashflowID)
    event.preventDefault()
    // add to RXJS stream processed with cashflowDelete.pipe above
    this.cashflowDelete$.next(cashflowID)

    // hide deleted CashFlow row in table
    const displayAmounts = this.cashflowAmountTargets;
    const cashflowRows = this.cashflowTableRowTargets;
    for (let r in cashflowRows) {
      let id = cashflowRows[r].getAttribute('data-cashflow-id')
      if (id == cashflowID) {
        cashflowRows[r].hidden = 1
        tableIdx = parseInt(r) - 1 // .hidden makes table one row smaller

        let oldAmountHTML = displayAmounts[r].innerHTML
        let amountStart = oldAmountHTML.search(/[$]/)
        let oldAmount = oldAmountHTML.slice(amountStart+1)
        // store adjust needed for Balance column
        adjustAmount = -1 * parseFloat(oldAmount)
      }
    }
    this.adjustBalances(tableIdx, adjustAmount)
  }

  // using click event
  actionEditAmount(event) {
    // hide value in table
    let displayAmount = event.currentTarget
    let cashflowID = displayAmount.getAttribute('data-cashflow-id')
    displayAmount.hidden = 1

    console.log("Stimulus[CASHFLOW]: actionEditAmount", cashflowID)

    // unhide input field so user can change value
    const inputAmounts = this.cashflowNewAmountTargets
    for (let i in inputAmounts) {
      let id = inputAmounts[i].getAttribute('data-cashflow-id')
      if (id == cashflowID) {
        inputAmounts[i].hidden = 0
        inputAmounts[i].focus()
        inputAmounts[i].select()
        break
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
    console.log("Stimulus[CASHFLOW]: keyCode is ", event.keyCode)

    // get new value and hide input field
    let inputAmount = event.currentTarget
    let cashflowID = inputAmount.getAttribute('data-cashflow-id')
    let newAmount = inputAmount.value
    let adjustAmount = 0
    let tableIdx = -1
    inputAmount.hidden = 1
    inputAmount.blur()

    if (send_amount) {
      console.log("Stimulus[CASHFLOW]: actionPutNewAmount", cashflowID, newAmount)
    }

    // unhide and update value in table (if changed)
    const displayAmounts = this.cashflowAmountTargets;
    const cashflowRows = this.cashflowTableRowTargets;
    for (let i in displayAmounts) {
      let id = displayAmounts[i].getAttribute('data-cashflow-id')

      if (id == cashflowID) {
        tableIdx = parseInt(i)
        if (send_amount) {
          let oldAmountHTML = displayAmounts[i].innerHTML
          let amountStart = oldAmountHTML.search(/[$]/)
          let oldAmount = oldAmountHTML.slice(amountStart+1)
          // store adjust needed for Balance column
          adjustAmount = parseFloat(newAmount) - parseFloat(oldAmount)

          // test if c.Amount actually changed or not
          if (adjustAmount == 0) {
            send_amount = 0
          } else {
            // overwrite c.Amount
            displayAmounts[i].innerHTML = "$"+(parseFloat(newAmount).toFixed(2))
          }
        }
        displayAmounts[i].hidden = 0
        break
      }
    }

    if (send_amount) {
      this.adjustBalances(tableIdx, adjustAmount)

      // add to RXJS stream processed with cashflowPut.pipe
      this.cashflowPut$.next([cashflowID, "amount", newAmount]);
    }
  }
}
