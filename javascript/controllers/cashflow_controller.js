/*
 * SPDX-FileCopyrightText: 2022 Brian Welty
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static get targets() {
    return [ "cashflowAmount", "cashflowNewAmount" ]
  }

  connect() {
    console.log("Stimulus connected!", this.element);
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
        // how to auto select for input?
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

    // add to RXJS stream processed with cashflowPutAmount.pipe
  }
}
