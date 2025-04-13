import React, { useContext } from "react";

const AppContext = React.createContext();
export const useAppContext = () => useContext(AppContext);

export const AppProvider = ({ children }) => {
    return (
        <AppContext.Provider value={{  }}>
            { children }
        </AppContext.Provider>
    );
};