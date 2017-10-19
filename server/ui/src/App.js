import React, { Component } from 'react';
import logo from './logo.svg';
import './App.css';
import Terminal from "xterm";

class App extends Component {
  render() {
      return (
          <div>hello</div>
    );
  }
}

console.log("111hh...");
function init()
{
    const term = new Terminal();
    term.open(document.getElementById('term'));
    term.write("Hello from \\033[1;3;31mxterm.js\\033[0m $")
    console.log("hh...");
}

init();

export default App;
