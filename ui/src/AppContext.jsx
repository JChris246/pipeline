import React, { useContext, useState } from "react";

const AppContext = React.createContext();
export const useAppContext = () => useContext(AppContext);

export const AppProvider = ({ children }) => {
    const [pipelines, setPipelines] = useState([]);
    const [selectedPipeline, setSelectedPipeline] = useState("");
    const [showDetails, setShowDetails] = useState(false);

    return (
        <AppContext.Provider value={{ pipelines, setPipelines, selectedPipeline, setSelectedPipeline,
            showDetails, setShowDetails }}>
            { children }
        </AppContext.Provider>
    );
};