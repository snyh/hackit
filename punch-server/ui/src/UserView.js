import React, { Component } from 'react';
import XTerm from "./react-xterm.js";

import { Loader, Message, Container, Grid, Header, Divider } from 'semantic-ui-react';

class Status extends Component {
    constructor(props) {
        super(props)
        this.state = {
            uuid : "unknown",
        }
    }

    componentDidMount() {
        const tick = () => {
            fetch(`http://${this.props.ttyServer}/tty/status`).then( (resp) => {
                resp.json().then( id => {
                    this.setState({uuid: id});
                    this.timer = setTimeout(tick, 3000);
                })
            }).catch( () => {
                this.setState({uuid: "invalid"})
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
        const url = `http://localhost:3000/connect/${key}`
        return (
            <div>
                The magic key is <b>{key}</b>
                <p>
                    you can send the link below to your friend to hack your system
                    <Message>
                        <a href={url} target="_blank">{url}</a>
                    </Message>
                </p>
            </div>
        )
    }
}


const LocalStatusUnknown = "unknown"
const LocalStatusError = "error"
const LocalStatusConnected = "connected"

class UserView extends Component {
    constructor(props) {
        super(props)

        const port = this.props.match.params.port;

        this.state = {
            localStatus: LocalStatusUnknown,
            invalidSource: port === undefined
        }

        this.ttyServer = `127.0.0.1:${port}`
        const backend = new WebSocket(`ws://${this.ttyServer}/tty`)
        backend.onerror = this.toError.bind(this)
        backend.onopen = this.toConnected.bind(this, backend)
    }

    toError() {
        this.setState({
            localStatus: LocalStatusError
        })
    }
    toConnected(ws) {
        this.setState({
            localStatus: LocalStatusConnected,
            backend: ws
        })
    }

    render() {
        console.log("SSS:", this.state)
        switch (this.state.localStatus) {
            case LocalStatusError:
                return (<CreateView />);
            case LocalStatusConnected:
                return (<DetailView ttyServer={this.ttyServer} backend={this.state.backend}/>)
            default:
                return (<div>loading</div>)
        }
    }
}

class CreateView extends Component {
    render() {
        return <div>请使用hackit的客户端进入本页面</div>
    }
}

class DetailView extends Component {
    render() {
        return (
            <Container>
                <Header>The terminal is operated by the hacker !</Header>
                <Status ttyServer={this.props.ttyServer}/>
                <Divider/>
                <Grid divided>
                    <Grid.Row>
                        <Grid.Column height="600px">
                            <XTerm backend={this.props.backend} />
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

export default UserView;
