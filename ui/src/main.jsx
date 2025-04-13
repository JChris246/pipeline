import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";

import App from "./pages/App";
import Error from "./pages/Error";
import "./index.css";

import { AppProvider } from "./AppContext";
import { Notification, NotificationProvider } from "./components/Notification";

const router = createBrowserRouter([
    {
        path: "/",
        element: <App />,
        errorElement: <Error />,
    }
]);

createRoot(document.getElementById("root")).render(
    <StrictMode>
        <NotificationProvider>
            <Notification />
            <AppProvider>
                <RouterProvider router={router} />
            </AppProvider>
        </NotificationProvider>
    </StrictMode>
);
