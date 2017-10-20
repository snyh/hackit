import React, { Component } from 'react';
import './App.css';

import XTerm from "./react-xterm.js";

import 'semantic-ui-css/semantic.min.css';

import { Message, Container, Grid, Header, Divider } from 'semantic-ui-react';

class App extends Component {
    render() {
        return (
            <Container>
                <Header> The terminal is operated by the hacker </Header>
                <Divider/>
                <Grid divided>
                    <Grid.Row>
                        <Grid.Column height="600px">
                            <XTerm backend="ws://127.0.0.1:8080/tty" />
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
