import React, { useContext, useEffect, useRef, useState } from "react";
import { CheckCircleIcon, WarningIcon, InfoIcon, ErrorIcon } from "../icons";

const NotificationType = { Success: 0, Warning: 1, Info: 2, Error: 3 };

const NotificationContext = React.createContext();
const useNotificationContext = () => useContext(NotificationContext);

const NotificationProvider = ({ children }) => {
    const [notificationDialog, setNotificationDialog] = useState({});
    const defaultArgs = { open: false, message: "", type: NotificationType.Info, duration: 3500, cb: null };

    const display = args => setNotificationDialog({ ...defaultArgs, ...args, open: true });
    const close = () => setNotificationDialog({ ...defaultArgs, open: false });

    return (
        <NotificationContext.Provider value={{ display, close, notificationDialog }}>
            { children }
        </NotificationContext.Provider>
    );
};

const Notification = () => {
    const { notificationDialog: { open, message, type, duration, cb }, close: handleClose } = useNotificationContext();
    const componentRef = useRef();

    const close = () =>  {
        if (cb)
            cb();
        handleClose();
    };

    useEffect(() => {
        if (open) {
            let timeout = setTimeout(close, duration);
            componentRef.current?.addEventListener("mouseover", () => clearTimeout(timeout));
            componentRef.current?.addEventListener("mouseout", () => timeout = setTimeout(close, duration));

            return () => clearTimeout(timeout);
        }
    // I do as I like
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [open]);

    const icons = [
        () => <CheckCircleIcon className="h-4 w-4 text-green-500 mt-0.5" />,
        () => <WarningIcon className="h-4 w-4 text-orange-400 mt-0.5" />,
        () => <InfoIcon className="h-4 w-4 text-blue-500 mt-0.5" />,
        () => <ErrorIcon className="h-4 w-4 text-red-500 mt-0.5" />
    ];

    return (
        open ? <div ref={componentRef} className="max-w-sm bg-slate-800/95 backdrop-blur-xl border border-slate-700/50 rounded-xl shadow-2xl absolute bottom-4 left-4 z-50 transform transition-all duration-300" role="alert">
            <div className="flex p-4">
                <div className="flex-shrink-0">
                    {icons[type]()}
                </div>
                <div className="ml-3 flex-1">
                    <p className="text-sm text-slate-200 font-medium">
                        {message}
                    </p>
                </div>
            </div>
        </div> : <></>
    );
};

// I do as I like
// eslint-disable-next-line react-refresh/only-export-components
export { Notification, NotificationType, useNotificationContext, NotificationProvider };