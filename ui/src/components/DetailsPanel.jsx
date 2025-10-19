import { useEffect, useState } from "react";

import { request } from "../utils/Fetch";
import { useAppContext } from "../AppContext";
import { NotificationType, useNotificationContext } from "./Notification";
import { formatDate, getDuration } from "../utils/utils";
import { BackIcon, DocumentIcon } from "../icons";

import EmptyPipelineRuns from "./EmptyPipelineRuns";

const DetailsPanel = ({ goBack }) => {
    const [pipelineRuns, setPipelineRuns] = useState([]);
    const { selectedPipeline: pipeline } = useAppContext();
    const { display: displayNotification } = useNotificationContext();

    useEffect(() => {
        if (!pipeline || pipeline.trim().length < 1)
            return;

        request({ url: "/api/pipelines/" + pipeline + "/runs",
            callback: ({ msg, success, json }) => {
                if (success) {
                    setPipelineRuns(json);
                } else {
                    displayNotification({ message: "An error occurred fetching pipeline runs: " + msg, type: NotificationType.Error });
                }
            }
        });
    // I do as I like
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pipeline]);

    return (
        <div className="flex flex-col fixed md:static w-full lg:w-2/3 h-screen overflow-y-scroll
            bg-gradient-to-br from-slate-800 via-slate-850 to-slate-900 border-none lg:border-l border-slate-700/50">

            {/* Header with back button */}
            <div className="flex justify-between items-center p-6 border-b border-slate-700/50 backdrop-blur-sm bg-slate-800/50">
                <h1 className="text-2xl font-semibold text-slate-100">Pipeline Runs</h1>
                <button onClick={() => { if (goBack) goBack(); }}
                    className="rounded-xl px-4 py-2 bg-gradient-to-r from-amber-600 to-amber-500 hover:from-amber-500 hover:to-amber-400 text-white
                        font-medium shadow-lg hover:shadow-amber-500/25 transition-all duration-200 block md:hidden
                        focus:ring-2 focus:ring-amber-500/50">
                    <BackIcon className="w-4 h-4 inline mr-2" />
                    Back
                </button>
            </div>

            {/* Content area */}
            <div className="flex-1 px-4">
                { pipelineRuns.length < 1 ? <EmptyPipelineRuns/> : (
                    pipelineRuns.map((run, index) => {
                        /* Pipeline runs display */
                        return (
                            <div key={index} className="bg-slate-800/40 backdrop-blur-sm border border-slate-700/30 rounded-xl overflow-hidden
                                hover:bg-slate-700/40 transition-all duration-200 group my-4">
                                {/* Run Header */}
                                <div className="p-6 border-b border-slate-700/30 flex justify-between items-start">
                                    <div className="flex-1">
                                        <h4 className="text-lg font-semibold text-slate-200 mb-1">
                                            {/* Should this be reverse indexed? */}
                                            {`Run #${index + 1}`}
                                        </h4>
                                        <div className="flex items-center space-x-4 text-sm text-slate-400">
                                            <span>Started: {formatDate(run.startedAt)}</span>
                                            {run.endedAt && ( // if endedAt has value, we can calculate the duration
                                                <span>Duration: {getDuration(run.startedAt, run.endedAt)}</span>
                                            )}
                                        </div>
                                    </div>
                                    <div className={`px-3 py-1.5 rounded-full text-xs font-medium border ${
                                        run.successful
                                            ? "bg-emerald-500/20 text-emerald-400 border-emerald-500/30"
                                            : "bg-red-500/20 text-red-400 border-red-500/30"
                                    }`}>
                                        {run.successful ? "Success" : "Failed"}
                                    </div>
                                </div>

                                {/* Stages */}
                                <div className="p-4">
                                    <h5 className="text-sm font-medium text-slate-300 mb-3 flex items-center">
                                        <DocumentIcon className="w-4 h-4 mr-2" />
                                        {/* Should I instead count the number of successful runs? */}
                                        Stages ({run.stages.length})
                                    </h5>
                                    <div className="space-y-2">
                                        {run.stages.map((stage, stageIndex) => (
                                            <div key={stageIndex} className="flex items-center justify-between p-3 bg-slate-700/30 rounded-lg">
                                                <div className="flex items-center space-x-3">
                                                    <div className={`w-2 h-2 rounded-full ${stage.skipped ? "bg-slate-500" : stage.successful ?
                                                        "bg-emerald-500" : "bg-red-500"}`} />
                                                    <span className="text-sm font-medium text-slate-200">
                                                        {stage.taskName}
                                                    </span>
                                                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                                                        stage.skipped ? "bg-slate-600/50 text-slate-400" : stage.successful ?
                                                            "bg-emerald-600/20 text-emerald-400" : "bg-red-600/20 text-red-400"}`}>
                                                        {stage.skipped ? "Skipped" : stage.successful ? "Success" : "Failed"}
                                                    </span>
                                                </div>
                                                {!stage.skipped && stage.startedAt && stage.endedAt && (
                                                    <span className="text-xs text-slate-400">
                                                        {getDuration(stage.startedAt, stage.endedAt)}
                                                    </span>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            </div>
                        );
                    })
                )}
            </div>
        </div>
    );
};

export default DetailsPanel;