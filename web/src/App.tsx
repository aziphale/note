import React, {ChangeEvent, KeyboardEventHandler, useEffect, useState} from 'react';
import './App.css';
import Blank from "./part/Blank";
import transfer from "./utils/markdown";
import refresh from "./utils/highlight";

function App() {

    const [content, setContent] = useState('');

    useEffect(() => {
        console.log("init...")

        fetch("api/newest")
            .then(res => res.json())
            .then(res => setContent(res.data));

        let latest = new WebSocket(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/api/echo`);
        latest.onopen = () => {
            console.log("connect successfully");
        };
        latest.onclose = () => {
            alert(`lost connection on ${new Date().toLocaleString()}`);
        };
        latest.onmessage = (response) => {
            setContent(response.data)
        };
    }, []);

    useEffect(() => {
        console.log("refresh");
        refresh();
    })

    let onKeyPress: React.KeyboardEventHandler<HTMLTextAreaElement> = (event) => {
        if (event.code === "Enter") {
            fetch("api/update", {
                headers: {
                    "Content-Type": "text/plain"
                },
                method: "POST",
                body: content
            }).then(res => res.json()).then(res => {
                if (res.status && res.status === 200) {
                    console.log("updated");
                } else {
                    alert(`lost connection on ${new Date().toLocaleString()}`);
                }
            });
        }
    }

    let onChange: React.ChangeEventHandler<HTMLTextAreaElement> = (event: ChangeEvent<HTMLTextAreaElement>) => {
        setContent(event.target.value);
    }

    return (
        <div className="App">
            <textarea id='note-md-id' defaultValue={content} onChange={onChange} onKeyDown={onKeyPress}></textarea>
            <Blank/>
            <div id='note-html-id' dangerouslySetInnerHTML={transfer(content)}></div>
        </div>
    );
}

export default App;
