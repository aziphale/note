import React, {ChangeEvent, useEffect, useState} from 'react';
import './App.css';
import transfer from "./utils/markdown";
import refresh from "./utils/highlight";

const INDENT = '    ';
const CURRENT_CONTENT = '\n<a id="current"/>\n';
const CURRENT_TAG = '#current';

const enum SHOW_MODE {
    HALF,
    EDIT,
    SHOW
}

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
            location.reload();
        };
        latest.onmessage = (response) => {
            setContent(response.data)
        };

        setInterval(() => {
            latest.send("ping");
            console.log("websocket alive");
        }, 30 * 1000);
    }, []);

    useEffect(() => {
        console.log("refresh");
        refresh();
    }, [content]);

    let onKeyPress: React.KeyboardEventHandler<HTMLTextAreaElement> = (event) => {
        let markdown: HTMLTextAreaElement = event.target as HTMLTextAreaElement;

        // tab => four blanks
        if (event.code === "Tab") {
            event.preventDefault();
            let start = markdown.selectionStart;
            let end = markdown.selectionEnd;
            markdown.value = markdown.value.substring(0,start) + INDENT + markdown.value.substring(end);
        }

        // update content and refresh preview
        if (event.code === "Enter") {
            fetch("api/update", {
                headers: {
                    "Content-Type": "text/plain"
                },
                method: "POST",
                body: markdown.value
            }).then(res => res.json()).then(res => {
                if (res.status && res.status === 200) {
                    console.log("updated");
                } else {
                    alert(`lost connection on ${new Date().toLocaleString()}`);
                }
            });
            setContent(markdown.value);
        }
    }

    let onChange: React.ChangeEventHandler<HTMLTextAreaElement> = (event: ChangeEvent<HTMLTextAreaElement>) => {

    }

    const [mode, setMode] = useState(SHOW_MODE.HALF);

    useEffect(() => {
        // @ts-ignore
        let viewChange = event => {
            if (event.code === "F11") {
                event.preventDefault();
                console.log("half view");
                setMode(SHOW_MODE.HALF);
            }
            if (event.code === "F10") {
                event.preventDefault();
                console.log("show view");
                setMode(SHOW_MODE.SHOW);
            }
            if (event.code === "F9") {
                event.preventDefault();
                console.log("edit view");
                setMode(SHOW_MODE.EDIT);
            }
        }
        addEventListener("keydown", viewChange);
        return () => {
            removeEventListener("keydown", viewChange);
        }
    }, [mode]);

    console.log("view init")

    return (
        <div className="App">
            <textarea id='note-md-id' className={SHOW_MODE.HALF === mode
                ? "half" : (SHOW_MODE.EDIT === mode ? "full" : "hide")}
                      defaultValue={content}
                      onChange={onChange}
                      onKeyDown={onKeyPress}></textarea>
            <div id='note-blank-id' className={SHOW_MODE.HALF === mode ? "blank" : "hide"}></div>
            <div id='note-html-id' className={SHOW_MODE.HALF === mode
                ? "half" : (SHOW_MODE.SHOW === mode ? "full" : "hide")}
                 dangerouslySetInnerHTML={transfer(content)}></div>
        </div>
    );
}

export default App;
