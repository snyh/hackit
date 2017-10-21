import "xterm/dist/xterm.css";

const React = require("react");
const Xterm = require("xterm");
const className = require('classnames');

Xterm.loadAddon('attach');
Xterm.loadAddon('fullscreen');

class XTerm extends React.Component {
    constructor(props, context) {
        super(props, context);
        this.onInput = (data) => {
            this.props.onInput && this.props.onInput(data);
        };
        this.fit = () => {
            var geometry = this.proposeGeometry(this.xterm);
            this.resize(geometry.cols, geometry.rows);
        };
        this.state = {
            isFocused: false
        };
    }
    getXTermInstance() {
        if (!this.xtermInstance) {
            this.xtermInstance = this.props.xtermInstance || Xterm;
        }
        return this.xtermInstance;
    }
    attachAddon(addon) {
        addon.attach(this.xtermInstance);
    }
    componentDidMount() {
        const xtermInstance = this.getXTermInstance();
        this.xterm = new xtermInstance(this.props.options);
        this.xterm.open(this.refs.container, true);
        this.xterm.on('focus', this.focusChanged.bind(this, true));
        this.xterm.on('blur', this.focusChanged.bind(this, false));
        if (this.props.onInput) {
            this.xterm.on('data', this.onInput);
        }
        if (this.props.value) {
            this.xterm.write(this.props.value);
        }

        const backend = this.props.backend;
        if (backend !== undefined) {
            this.backendWS = new WebSocket(backend)
            this.xterm.attach(this.backendWS);
        } else {
            this.writeln("Please setup the backend address.")
        }
    }
    componentWillUnmount() {
        if (this.xterm) {
            this.xterm.destroy();
            this.xterm = null;
        }
        if (this.backendWS) {
            this.backendWS.close()
        }
    }
    getXTerm() {
        return this.xterm;
    }
    write(data) {
        this.xterm.write(data);
    }
    writeln(data) {
        this.xterm.writeln(data);
    }
    focus() {
        if (this.xterm) {
            this.xterm.focus();
        }
    }
    focusChanged(focused) {
        this.setState({
            isFocused: focused,
        });
        this.props.onFocusChange && this.props.onFocusChange(focused);
    }
    resize(cols, rows) {
        this.xterm.resize(Math.round(cols), Math.round(rows));
    }
    setCursorBlink(blink) {
        if (this.xterm && this.xterm.cursorBlink !== blink) {
            this.xterm.cursorBlink = blink;
            this.xterm.refresh(0, this.xterm.rows - 1);
        }
    }
    proposeGeometry(term) {
        var parentElementStyle = window.getComputedStyle(term.element.parentElement), parentElementHeight = parseInt(parentElementStyle.getPropertyValue('height')), parentElementWidth = Math.max(0, parseInt(parentElementStyle.getPropertyValue('width')) - 17), elementStyle = window.getComputedStyle(term.element), elementPaddingVer = parseInt(elementStyle.getPropertyValue('padding-top')) + parseInt(elementStyle.getPropertyValue('padding-bottom')), elementPaddingHor = parseInt(elementStyle.getPropertyValue('padding-right')) + parseInt(elementStyle.getPropertyValue('padding-left')), availableHeight = parentElementHeight - elementPaddingVer, availableWidth = parentElementWidth - elementPaddingHor, container = term.rowContainer, subjectRow = term.rowContainer.firstElementChild, contentBuffer = subjectRow.innerHTML, characterHeight, rows, characterWidth, cols, geometry;
        subjectRow.style.display = 'inline';
        subjectRow.innerHTML = 'W';
        characterWidth = subjectRow.getBoundingClientRect().width;
        subjectRow.style.display = '';
        characterHeight = parseFloat(subjectRow.offsetHeight);
        subjectRow.innerHTML = contentBuffer;
        rows = availableHeight / characterHeight;
        cols = availableWidth / characterWidth;
        geometry = { cols: cols, rows: rows };
        return geometry;
    }
    ;
    render() {
        const terminalClassName = className('ReactXTerm', this.state.isFocused ? 'ReactXTerm--focused' : null, this.props.className);
        return (React.createElement("div", { ref: "container", className: terminalClassName }));
    }
}

export default XTerm;
