/*
This shows a line chart of the data given
*/

import React, {PropTypes} from 'react';
import DataTransformUpdater from './DataUpdater';

import {Line} from 'react-chartjs';
import moment from 'moment';

class LineChart extends DataTransformUpdater {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired,
        transform: PropTypes.string
    }

    // transformDataset is required for DataUpdater to set up the modified state data
    transformDataset(d) {
        let dataset = new Array(d.length);

        for (let i = 0; i < d.length; i++) {
            dataset[i] = {
                x: moment.unix(d[i].t),
                y: d[i].d
            }
        }

        // Now return the data necessary to use the line chart
        return {
            datasets: [
                {
                    label: name,
                    data: dataset,
                    lineTension: 0
                }
            ]
        };
    }

    render() {
        return (<Line data={this.data} options={{
            legend: {
                display: false
            },
            scales: {
                xAxes: [
                    {
                        type: 'time',
                        position: 'bottom'
                    }
                ]
            }
        }}/>);
    }
}

export default LineChart;

// generate creates a new view that displays a line chart. The view object is set up
// so that it is totally ready to be passed as a result of the shower function
export function generateLineChart(transform) {
    let component = LineChart;

    // If we're given a transform, wrap the LineChart so that we can pass transform into the class.
    if (transform != null) {
        component = React.createClass({
            render: function() {
                return (<LineChart {...this.props} transform={transform}/>);
            }
        });
    }

    return {initialState: {}, component: component, width: "expandable-half"};
}
