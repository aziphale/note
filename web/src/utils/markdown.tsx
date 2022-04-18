import showdown from 'showdown';

const converter = new showdown.Converter({ tables: true });

interface htmlContent {
    __html: string
}

function transfer(markdown: string): htmlContent {
    return { __html: (converter.makeHtml(markdown) || '<h1>preview here</h1>')};
}

export default transfer;