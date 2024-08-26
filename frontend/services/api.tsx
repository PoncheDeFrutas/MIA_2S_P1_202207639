
export async function POST<T>(path: string, content: T): Promise<any> {
    console.log(content);
    const response = await fetch(path, {
        method: 'POST',
        body: JSON.stringify(content),
        headers: {
            'Content-Type': 'application/json',
        },
    });
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
}

export async function GET(path: string): Promise<any> {
    const response = await fetch(path, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    });
    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json();
}
