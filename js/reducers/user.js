// The userReducer maintains state for all users - each user page that was visited in this session has
// its state maintained here. Each action that operates on a user contains a "name" field, which states the user
// in whose context to operate

// The keys in this object are going to be the user names
const InitialState = {};

import userViewReducer, {UserViewInitialState} from './userView';
import userEditReducer, {UserEditInitialState} from './userEdit';
import deviceCreateReducer, {DeviceCreateInitialState} from './deviceCreate';

// The initial state of a specific user
const UserInitialState = {
    edit: UserEditInitialState,
    view: UserViewInitialState,
    create: DeviceCreateInitialState
};

export default function userReducer(state = InitialState, action) {
    if (!action.type.startsWith("USER_"))
        return state;

    // Set up the new state
    let newState = {
        ...state
    };

    // If the user already has a state, copy the next level, otherwise, initialize the next level
    if (state[action.uname] !== undefined) {
        newState[action.uname] = {
            ...state[action.uname]
        }
    } else {
        newState[action.uname] = Object.assign({}, UserInitialState);
    }

    // Now route to the appropriate reducer
    if (action.type.startsWith("USER_EDIT_"))
        newState[action.uname].edit = userEditReducer(newState[action.uname].edit, action);
    if (action.type.startsWith("USER_VIEW_"))
        newState[action.uname].view = userViewReducer(newState[action.uname].view, action);
    if (action.type.startsWith("USER_CREATEDEVICE_"))
        newState[action.uname].create = deviceCreateReducer(newState[action.uname].create, action);

    return newState;
}

// get the user page from the state - the state might not have this
// particular page initialized, meaning that it wasn't acted upon
export function getUserState(user, state) {
    return (state.user[user] !== undefined
        ? state.user[user]
        : UserInitialState);
}
