import { DocumentIcon } from "../icons";

const EmptyPipelineRuns = () => {
    return (
        <div className="flex flex-col items-center justify-center h-full text-center">
            <div className="w-24 h-24 bg-slate-700 rounded-full flex items-center justify-center mb-6">
                <DocumentIcon className="w-12 h-12 text-slate-400" />
            </div>
            <h3 className="text-2xl font-semibold text-slate-300 mb-2">No Pipeline Runs</h3>
            <p className="text-slate-400">This pipeline hasn't been executed yet</p>
        </div>
    );
};

export default EmptyPipelineRuns;