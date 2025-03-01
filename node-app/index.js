const express = require('express');
const dotenv = require('dotenv');

dotenv.config();


const app = express();

app.get('/', (_, res) => {
    console.log("Hello World!");
    console.warn('This is a warning!');
    console.error('This is an error!');
    res.status(200).json({ message: 'Hello World!' });
});

app.listen(8080, () => {
    console.log('Server is running on port 8080');
});