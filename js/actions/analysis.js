
export function setState(state) {
    return { type: "ANALYSIS_STATE", value: state };
}

export function setDatasetState(key, state) {
    return { type: "DATASET_STATE", value: state, key: key };
}

export function clear() {
    return { type: "ANALYSIS_CLEAR" };
}

export function query() {
    return { type: "DATASET_QUERY" };  // This triggers the analysis saga, which will query for dataset.
}