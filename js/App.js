import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {Router, Route, IndexRoute, browserHistory} from 'react-router'

import Theme from './Theme';

import Main from './Main';
import User from './User';
import Device from './Device';
import Stream from './Stream';

class App extends Component {
    static propTypes = {
        history: PropTypes.object.isRequired
    };
    render() {
        return (
            <Router history={this.props.history}>
                <Route path="/" component={Theme}>
                    <IndexRoute component={Main}/>
                    <Route path="/:user" component={User}/>
                    <Route path="/:user/:device" component={Device}/>
                    <Route path="/:user/:device/:stream" component={Stream}/>
                </Route>
            </Router>
        );
    }
}

export default App;
