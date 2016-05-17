import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {editCancel, go, showMessage} from '../actions';

import Dialog from 'material-ui/Dialog';
import FlatButton from 'material-ui/FlatButton';
import {RadioButton, RadioButtonGroup} from 'material-ui/RadioButton';
import TextField from 'material-ui/TextField';
import Checkbox from 'material-ui/Checkbox';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import Avatar from 'material-ui/Avatar';
import Snackbar from 'material-ui/Snackbar';

import storage from '../storage';

class UserEdit extends Component {
    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired,
        state: PropTypes.object.isRequired,
        roles: PropTypes.object.isRequired,
        onCancelClick: PropTypes.func.isRequired,
        nicknameChange: PropTypes.func.isRequired,
        emailChange: PropTypes.func.isRequired,
        publicChange: PropTypes.func.isRequired,
        descriptionChange: PropTypes.func.isRequired,
        roleChange: PropTypes.func.isRequired,
        passwordChange: PropTypes.func.isRequired,
        password2Change: PropTypes.func.isRequired,
        onDelete: PropTypes.func.isRequired,
        onSave: PropTypes.func.isRequired
    }
    constructor(props) {
        super(props);
        this.state = {
            dialogopen: false,
            message: ""
        };
    }
    dialogDelete() {
        // Delete the user
        this.setState({dialogopen: false, message: "Deleting user..."});
        storage.del(this.props.user.name).then((result) => {
            if (result == "ok")
                this.props.onDelete();
            else {
                this.setState({message: result.msg});
            }

        }).catch((err) => {
            console.log(err);
            this.setState({message: "Failed to delete user"});
        });
    }
    save() {
        let state = Object.assign({}, this.props.state);
        if (state.password !== undefined) {
            if (state.password != state.password2) {
                this.setState({message: "Passwords do not match"});
                return;
            }
            if (state.password == "") {
                delete state.password;

            }
            delete state.password2;
        }

        // Now delete any state values that match current values
        Object.keys(state).forEach((key) => {
            if (this.props.user[key] !== undefined) {
                if (this.props.user[key] == state[key]) {
                    delete state[key];
                }
            }
        });
        this.setState({message: "Updating user..."});

        // Finally, update the user
        storage.update(this.props.user.name, state).then((result) => {
            if (result.ref === undefined) {
                this.props.onSave();
                return;
            }
            this.setState({message: result.msg});
        }).catch((err) => {
            console.log(err);
            this.setState({message: "Failed to update user"});
        });

    }
    render() {
        let user = this.props.user;
        let edits = this.props.state;
        let nickname = user.name;
        if (user.nickname !== undefined && user.nickname != "") {
            nickname = user.nickname;
        }
        if (edits.nickname !== undefined && edits.nickname != "") {
            nickname = edits.nickname;
        }
        return (
            <Card style={{
                textAlign: "left"
            }}>
                <CardHeader title={nickname} subtitle={user.name} avatar={< Avatar > U < /Avatar>}/>
                <CardText>
                    <TextField hintText="Nickname" floatingLabelText="Nickname" value={edits.nickname !== undefined
                        ? edits.nickname
                        : user.nickname} onChange={this.props.nicknameChange}/><br/>
                    <TextField hintText="Email" floatingLabelText="Email" value={edits.email !== undefined
                        ? edits.email
                        : user.email} onChange={this.props.emailChange}/><br/>
                    <h3>Public</h3>
                    <p>Whether or not the user can be accessed (viewed) by other users.</p>
                    <Checkbox label="Public" checked={edits.public !== undefined
                        ? edits.public
                        : user.public} onCheck={this.props.publicChange}/>
                    <h3>Description</h3>
                    <p>A user's description can be thought of as a README for the user.</p>
                    <TextField hintText="I am pretty awesome" floatingLabelText="Description" multiLine={true} fullWidth={true} value={edits.description !== undefined
                        ? edits.description
                        : user.description} style={{
                        marginTop: "-20px"
                    }} onChange={this.props.descriptionChange}/><br/>
                    <h3>Role</h3>
                    <p>A user's role determines the permissions given to operate upon ConnectorDB.</p>
                    <RadioButtonGroup name="role" valueSelected={edits.role !== undefined
                        ? edits.role
                        : user.role} onChange={this.props.roleChange}>
                        {Object.keys(this.props.roles).map((key) => (<RadioButton value={key} key={key} label={key + " - " + this.props.roles[key].description}/>))}
                    </RadioButtonGroup>
                    <h3>Password</h3>
                    <p>Change your user's password</p>
                    <TextField hintText="Type New Password" type="password" style={{
                        marginTop: "-30px"
                    }} value={edits.password !== undefined
                        ? edits.password
                        : ""} onChange={this.props.passwordChange}/>
                    <br/> {edits.password !== undefined
                        ? (<TextField hintText="Type New Password" type="password" floatingLabelText="Repeat New Password" value={edits.password2 !== undefined
                            ? edits.password2
                            : ""} onChange={this.props.password2Change}/>)
                        : null}
                    <br/>
                    <div style={{
                        paddingTop: "20px"
                    }}>
                        <FlatButton primary={true} label="Save" onTouchTap={() => this.save()}/>
                        <FlatButton label=" Cancel" onTouchTap={this.props.onCancelClick}/>
                        <FlatButton label="Delete" style={{
                            color: "red",
                            float: "right"
                        }} onTouchTap={() => this.setState({dialogopen: true})}/>
                    </div>
                </CardText>
                <Dialog title="Delete User" actions={[(<FlatButton label="Cancel" onTouchTap={() => this.setState({dialogopen: false})} keyboardFocused={true}/>), (<FlatButton label="Delete" onTouchTap={() => this.dialogDelete()}/>)]} modal={false} open={this.state.dialogopen}>
                    Are you sure you want to delete the user "{user.name}"?
                </Dialog>
                <Snackbar open={this.state.message != ""} message={this.state.message}/>
            </Card>
        );
    }
}

// It would be horrible to have all of these actions upstream - so we do them here.
export default connect((store) => ({roles: store.site.roles.user}), (dispatch, props) => ({
    nicknameChange: (e, txt) => dispatch({type: "USER_EDIT_NICKNAME", name: props.user.name, value: txt}),
    descriptionChange: (e, txt) => dispatch({type: "USER_EDIT_DESCRIPTION", name: props.user.name, value: txt}),
    passwordChange: (e, txt) => dispatch({type: "USER_EDIT_PASSWORD", name: props.user.name, value: txt}),
    password2Change: (e, txt) => dispatch({type: "USER_EDIT_PASSWORD2", name: props.user.name, value: txt}),
    roleChange: (e, role) => dispatch({type: "USER_EDIT_ROLE", name: props.user.name, value: role}),
    publicChange: (e, val) => dispatch({type: "USER_EDIT_PUBLIC", name: props.user.name, value: val}),
    emailChange: (e, val) => dispatch({type: "USER_EDIT_EMAIL", name: props.user.name, value: val}),
    onCancelClick: () => dispatch(editCancel("USER", props.user.name)),
    onSave: () => {
        dispatch(showMessage("Updated User"));
        dispatch(editCancel("USER", props.user.name));
    },
    onDelete: () => {
        dispatch(showMessage("User Deleted"));
        dispatch(go(""));
    }
}))(UserEdit);
