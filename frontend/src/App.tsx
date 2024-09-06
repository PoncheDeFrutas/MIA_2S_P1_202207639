import './App.css';
import React, {useState} from 'react';
import {EditorComponent} from '../components/editor';
import {POST} from '../services/api.tsx';

function App() {
    const [fileContent, setFileContent] = useState("");
    const [fileName, setFileName] = useState("");
    const [output, setOutput] = useState("");

    const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0];
        if (file && file.name.endsWith(".smia")) {
            setFileName(file.name);
            const reader = new FileReader();
            reader.onload = (e) => {
                setFileContent(e.target?.result as string);
            };
            reader.readAsText(file);
        } else {
            alert("Invalid file format. Please upload a .smia file.");
        }
    };

    const handleSaveFile = () => {
        const blob = new Blob([fileContent], {type: 'text/plain;charset=utf-8'});
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = 'file.smia';
        link.click();
        URL.revokeObjectURL(url);
    };

    const handleExecute = async () => {
        try {
            const response = await POST('http://localhost:5000/execute', {content: fileContent});
            const resultText = response.result;
            setOutput(resultText);
        } catch (error) {
            setOutput('Error: Unable to process the request.');
        }
    };

    const handleContentChange = (content: string) => {
        setFileContent(content);
    };

    return (
        <div className="App">
            <header className="App-header">
                <div className="header-left">
                    <h2>FruitPunchFS</h2>
                </div>
                <div className="header-right">
                    <div className="header-info">
                        <h2>{fileName || "No file selected"}</h2>
                    </div>
                    <div className="header-buttons">
                        <div className="file-upload-wrapper">
                            <input
                                type="file"
                                id="file-upload"
                                accept=".smia"
                                onChange={handleFileUpload}
                                className="file-input"
                            />
                            <button className="header-button custom-file-upload"
                                    onClick={() => document.getElementById('file-upload')?.click()}>
                                Upload File
                            </button>
                        </div>
                        <button className="header-button" onClick={handleExecute}>Execute</button>
                        <button className="header-button" onClick={handleSaveFile}>Save</button>
                    </div>
                </div>
            </header>
            <div className="content">
                <div className="left-side">
                    <EditorComponent
                        content={fileContent}
                        onContentChange={handleContentChange}
                    />
                </div>
                <div className="right-side">
                    <textarea
                        className="output-area"
                        readOnly
                        value={output}
                        placeholder="Output will appear here..."
                    ></textarea>
                </div>
            </div>
        </div>
    );
}

export default App;
