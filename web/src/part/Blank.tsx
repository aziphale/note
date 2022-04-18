import React from "react";

class Blank extends React.Component {

    constructor(props: {}) {
        super(props);
        this.state = {};
    }

    render() {
        return (
            <div id='note-blank-id'></div>
        );
    }
}

export default Blank;
