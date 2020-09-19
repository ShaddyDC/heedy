import TimeseriesInjector from "./worker/injector.js";

import datatableAnalyzer from "./worker/analyzers/datatable.js";
import linechartAnalyzer from "./worker/analyzers/linechart.js";
import correlationAnalyzer from "./worker/analyzers/correlation.js";
import summaryAnalyzer from "./worker/analyzers/summary.js";

import chartjsPreprocessor from "./worker/preprocessors/chartjs.js";
import datatablePreprocessor from "./worker/preprocessors/datatable.js";
import tablePreprocessor from "./worker/preprocessors/table.js";
/*

import insert from "./worker/preprocessors/insert.js";
import linechart from "./worker/preprocessors/linechart.js";
import dayview from "./worker/preprocessors/dayview.js";
*/
function setup(wkr) {
  console.log("timeseries_worker: starting");

  wkr.inject("timeseries", new TimeseriesInjector(wkr));

  wkr.timeseries.addAnalyzer(datatableAnalyzer);
  wkr.timeseries.addAnalyzer(linechartAnalyzer);
  wkr.timeseries.addAnalyzer(correlationAnalyzer);
  wkr.timeseries.addAnalyzer(summaryAnalyzer);

  wkr.timeseries.addPreprocessor("chartjs", chartjsPreprocessor);
  wkr.timeseries.addPreprocessor("datatable", datatablePreprocessor);
  wkr.timeseries.addPreprocessor("table", tablePreprocessor);
  /*
  wkr.timeseries.addPreprocessor("datatable", datatable);
  wkr.timeseries.addPreprocessor("insert", insert);
  wkr.timeseries.addPreprocessor("linechart", linechart);
  wkr.timeseries.addPreprocessor("dayview", dayview);
  */
}

export default setup;
