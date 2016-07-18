import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {createCancel, createObject} from '../actions';

import ObjectCreate from '../components/ObjectCreate';

import DownlinkEditor from '../components/edit/DownlinkEditor';
import EphemeralEditor from '../components/edit/EphemeralEditor';
import DatatypeEditor from '../components/edit/DatatypeEditor';
import SchemaEditor from '../components/edit/SchemaEditor';

class StreamCreate extends Component {
    static propTypes = {
        user: PropTypes.object.isRequired,
        device: PropTypes.object.isRequired,
        state: PropTypes.object.isRequired,
        callbacks: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancel: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired
    }
    render() {
        let state = this.props.state;
        let callbacks = this.props.callbacks;
        return (
            <ObjectCreate type="stream" state={state} callbacks={callbacks} required={< SchemaEditor value = {
                state.schema
            }
            onChange = {
                callbacks.schemaChange
            } />} parentPath={this.props.user.name + "/" + this.props.device.name} onCancel={this.props.onCancel} onSave={this.props.onSave}>
                <DatatypeEditor value={state.datatype} schema={state.schema} onChange={callbacks.datatypeChange}/>
                <DownlinkEditor value={state.downlink} onChange={callbacks.downlinkChange}/>
                <EphemeralEditor value={state.ephemeral} onChange={callbacks.ephemeralChange}/>
            </ObjectCreate >

        );
    }
}

export default connect((state) => ({roles: state.site.roles.device}), (dispatch, props) => {
    let name = props.user.name + "/" + props.device.name;
    return {
        callbacks: {
            nameChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_NAME", name: name, value: txt}),
            nicknameChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_NICKNAME", name: name, value: txt}),
            descriptionChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_DESCRIPTION", name: name, value: txt}),
            ephemeralChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_EPHEMERAL", name: name, value: txt}),
            downlinkChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_DOWNLINK", name: name, value: txt}),
            datatypeChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_DATATYPE", name: name, value: txt}),
            schemaChange: (e, txt) => dispatch({type: "DEVICE_CREATESTREAM_SCHEMA", name: name, value: txt})
        },
        onCancel: () => dispatch(createCancel("DEVICE", "STREAM", name)),
        onSave: () => dispatch(createObject("device", "stream", name, props.state))
    }
})(StreamCreate);
