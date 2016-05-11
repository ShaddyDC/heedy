// Actions are things that can happen... To make it happen, run store.dispatch(action())
import {push} from 'react-router-redux'

// getPage is a cooler version of navigator, which allows us to navigate to illegal pages such as logout
export function showPage(name) {
    // First, deal with special values
    switch (name) {
        case "logout":
            window.href = "/logout"
            return null;

    }
    // otherwise, go to the given page
    return push(name);
}

// set the search bar text
export function setSearchText(text) {
    return {type: 'SET_SEARCH_TEXT', value: text};
}
