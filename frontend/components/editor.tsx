import React from 'react';
import Editor from '@monaco-editor/react';
import './EditorComponent.css';

interface EditorComponentProps {
    content: string;
    onContentChange: (content: string) => void;
}

export class EditorComponent extends React.Component<EditorComponentProps> {
    handleChange = (value: string | undefined) => {
        if (value !== undefined) {
            this.props.onContentChange(value);
        }
    };

    render() {
        const { content } = this.props;
        return (
            <div className="editor-container">
                <Editor
                    theme="vs-dark"
                    language="typescript"
                    value={content}
                    options={{
                        minimap: { enabled: false },
                        wordWrap: 'on',
                    }}
                    onChange={this.handleChange}
                />
            </div>
        );
    }
}
