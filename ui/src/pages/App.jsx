import "./App.css";

import PipelineListPanel from "../components/PipelineListPanel";
import DetailsPanel from "../components/DetailsPanel";

import { useAppContext } from "../AppContext";

function App() {
    const { showDetails, setShowDetails, setSelectedPipeline } = useAppContext();

    const clearDetailPanel = () => {
        setShowDetails(false);
        setSelectedPipeline("");
    };

    return (

        <div className="flex">
            <PipelineListPanel/>
            { showDetails && <DetailsPanel goBack={clearDetailPanel}/> }
        </div>
    );
}

export default App;
