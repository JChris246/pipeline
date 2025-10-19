import { PipelineIcon, AddIcon } from "../icons";

const PipelineEmptyState = ({ searchFilter, onCreatePipeline }) => {
    return (
        <div className="flex-1 flex items-center justify-center p-8">
            <div className="text-center">
                <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-gradient-to-r from-slate-700 to-slate-600
                    flex items-center justify-center">
                    <PipelineIcon className="w-8 h-8 text-slate-400" />
                </div>
                <h3 className="text-lg font-medium text-slate-300 mb-2">No pipelines found</h3>
                <p className="text-sm text-slate-400 mb-4">
                    {searchFilter ? `No pipelines match "${searchFilter}"` : "Get started by creating your first pipeline"}
                </p>
                {!searchFilter ? (
                    <button onClick={onCreatePipeline} className="inline-flex items-center space-x-2 px-4 py-2 bg-gradient-to-r
                        from-emerald-600 to-emerald-500 hover:from-emerald-500 hover:to-emerald-400 text-white rounded-lg
                        font-medium transition-all duration-200 shadow-lg hover:shadow-emerald-500/25">
                        <AddIcon className="w-4 h-4" />
                        <span>Create Pipeline</span>
                    </button>
                ) : ""}
            </div>
        </div>
    );
};

export default PipelineEmptyState;
