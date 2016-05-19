const InitialState = {
    // roles represents the possible permissions allowed by ConnectorDB.
    // Note that the values given here correspond to the default ConnectorDB settings.
    // One can change ConnectorDB to have a different permissions structure, which would
    // make these values inconsistent with ConnectorDB.... so don't do that
    roles: {
        user: {
            user: {
                description: "can read/edit own devices and read public users/devices"
            },
            admin: {
                description: "has administrative access to the database"
            }
        },
        device: {
            user: {
                description: "has all permissions that the owning user has (Create/Read/Update/Delete)"
            },
            writer: {
                description: "can read and update basic properties of streams and devices"
            },
            reader: {
                description: "can read properties of streams and devices, and read their data"
            },
            none: {
                description: "the device is isolated - it only has access to itself and its own streams"
            }
        }
    },

    // navigation is displayed in the app's main nmenu
    navigation: [
        {
            title: "Progress Log",
            subtitle: "Manually insert data",
            icon: "star",
            page: "/"
        }, {
            title: "Profile",
            subtitle: "View your devices",
            icon: "face",
            page: "/{self}"
        }, {
            title: "Log Out",
            subtitle: "Exit your session",
            icon: "power_settings_new",
            page: "/logout"
        }
    ],

    // The currently logged in user and device. This is set up immediately on app start.
    // even before the app is rendered. Note that these are NOT updated along with
    // the app storage - this is the initial user and device
    thisUser: null,
    thisDevice: null,

    // The URL of the website, also available as global variable "SiteURL". This is set up
    // from the context on app load
    siteURL: "",

    // The status message to show in the snack bar
    status: "",
    statusvisible: false

};

export default function siteReducer(state = InitialState, action) {
    switch (action.type) {
        case 'LOAD_CONTEXT':
            var out = Object.assign({}, state, {
                siteURL: action.value.SiteURL,
                thisUser: action.value.ThisUser,
                thisDevice: action.value.ThisDevice
            });

            // now set up the navigation correctly (replace {self} with user name)
            for (var i = 0; i < out.navigation.length; i++) {
                out.navigation[i].title = out.navigation[i].title.replace("{self}", out.thisUser.name);
                out.navigation[i].subtitle = out.navigation[i].subtitle.replace("{self}", out.thisUser.name);
                out.navigation[i].page = out.navigation[i].page.replace("{self}", out.thisUser.name);
            }
            return out;
        case 'STATUS_HIDE':
            return {
                ...state,
                statusvisible: false
            };
        case 'SHOW_STATUS':
            return {
                ...state,
                statusvisible: true,
                status: action.value
            }
    }
    return state
}
