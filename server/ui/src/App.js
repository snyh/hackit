import React, { Component } from 'react';
import './App.css';

import XTerm from "./react-xterm.js";

import 'semantic-ui-css/semantic.min.css';

import { Message, Container, Grid, Header, Divider } from 'semantic-ui-react';


const TTY_SERVER = "127.0.0.1:8080"

class Status extends Component {
    constructor(props) {
        super(props)
        this.state = {
            uuid : "unknown"
        }
    }

    componentDidMount() {
        const tick = () => {
            fetch(`http://${TTY_SERVER}/tty/status`).then( (resp) => {
                resp.json().then( id => {
                    this.setState({uuid: id});
                    this.timer = setTimeout(tick, 3000);
                })
            }).catch( () => {
                this.setState({uuid: "unknown"})
                this.timer = setTimeout(tick, 3000);
            })
        };
        tick();
    }

    componentWillUnmount() {
        clearTimeout(this.timer);
    }

    render() {
        const key = this.state.uuid
        const url = `http://localhost:3000/connect/${key}?open`
        return (
            <div>
                The magic key is <b>{key}</b>
                <p>
                    you can tell your hacker to hack you from deepin system
                    <Message>
                        <code>ssh -T {key}@hackit.snyh.org</code>.
                        or
                        <code>./client {key}</code>
                    </Message>
                    or directly from web browser
                    <Message>
                        <a href={url} target="_blank">{url}</a>
                    </Message>
                </p>
            </div>
        )
    }
}

class App extends Component {
    render() {
        return (
            <Container>
                <Header>The terminal is operated by the hacker</Header>
                <Status/>
                <Divider/>
                <Grid divided>
                    <Grid.Row>
                        <Grid.Column height="600px">
                            <XTerm backend={`ws://${TTY_SERVER}/tty`} />
                        </Grid.Column>
                    </Grid.Row>
                    <Grid.Row>
                        <Grid.Column>
                            <Message>
                                <p><span>Hacker:</span> 123hhhhh chat message........</p>
                                <p><span>You:</span> hhhhh chat message........</p>
                            </Message>
                        </Grid.Column>
                    </Grid.Row>
                </Grid>
            </Container>
        );
    }
}

export default App;
