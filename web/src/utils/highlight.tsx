import hljs from 'highlight.js';
import 'highlight.js/styles/default.css'

function refresh(): void {
    hljs.highlightAll();
}

export default refresh;