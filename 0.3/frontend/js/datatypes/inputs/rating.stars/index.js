import React, { Component } from "react";
import PropTypes from "prop-types";

// TEMPORARY HACK: react-star-rating package is outdated on npm with an old react version.
// github has correct version, which needs to be compiled and stuff, so I just included the
// compiled files while waiting for an update to npm
// Once this is fixed, should also delete dependency classnames
//import StarRating from 'react-star-rating';
//import 'react-star-rating/dist/css/react-star-rating.min.css';
import StarRating from "./react-star-rating/react-star-rating.min";
import "./react-star-rating/react-star-rating.min.css";
import { addInput } from "../../datatypes";

import TimeChooser, { getTimestamp } from "../TimeChooser";

class StarInput extends Component {
  static propTypes = {
    state: PropTypes.object,
    setState: PropTypes.func,
    insert: PropTypes.func
  };
  render() {
    let value = this.props.state.value;
    if (value === undefined || value == null) value = 0;

    // rating={value} messes up our ability to set the rating again in current version of react star rating. We therefore can't have it set :(
    return (
      <StarRating
        name={this.props.path}
        totalStars={10}
        size={30}
        onRatingClick={(a, val) => {
          console.log("Changing value:", val);
          this.props.setState({ value: val["rating"] });
          this.props.insert(
            getTimestamp(this.props.state),
            val["rating"],
            false
          );
        }}
      />
    );
  }
}

// add the input to the input registry.
addInput("rating.stars", {
  width: "half",
  component: StarInput,
  style: {
    textAlign: "center"
  },
  dropdown: TimeChooser
});
