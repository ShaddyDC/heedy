export const DeviceCreateInitialState = {
    name: "",
    nickname: "",
    description: "",
    role: "none",
    public: false
};

export default function deviceCreateReducer(state, action) {
    switch (action.type) {
        case 'USER_CREATEDEVICE_CLEAR':
            return DeviceCreateInitialState;
        case 'USER_CREATEDEVICE_NAME':
            return {
                ...state,
                name: action.value
            };
        case 'USER_CREATEDEVICE_NICKNAME':
            return {
                ...state,
                nickname: action.value
            };
        case 'USER_CREATEDEVICE_DESCRIPTION':
            return {
                ...state,
                description: action.value
            };
        case 'USER_CREATEDEVICE_ROLE':
            return {
                ...state,
                role: action.value
            };
        case 'USER_CREATEDEVICE_PUBLIC':
            return {
                ...state,
                public: action.value
            };

    }
    return state;
}
