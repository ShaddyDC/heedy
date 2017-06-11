import React, { Component } from "react";
import PropTypes from "prop-types";
import TextField from "material-ui/TextField";

class NicknameEditor extends Component {
  static propTypes = {
    value: PropTypes.string.isRequired,
    onChange: PropTypes.func.isRequired,
    type: PropTypes.string.isRequired
  };

  render() {
    return (
      <div>
        <h3>Nickname</h3>
        <p>An easy to read title for your {this.props.type}</p>
        <TextField
          hintText="Nickname"
          floatingLabelText="Nickname"
          style={{
            marginTop: "-20px"
          }}
          value={this.props.value}
          onChange={this.props.onChange}
        />
        <br />
      </div>
    );
  }
}
export default NicknameEditor;
