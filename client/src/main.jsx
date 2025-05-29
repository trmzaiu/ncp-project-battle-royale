import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import App from "./App.jsx";
import "./index.css";
import process from 'process'

ReactDOM.createRoot(document.getElementById("root")).render(
    <BrowserRouter basename={process.env.NODE_ENV === 'production' ? "/royaka-2025-fe" : "/"}>
        <App />
    </BrowserRouter>
);