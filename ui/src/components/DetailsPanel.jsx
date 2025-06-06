import { useEffect, useState } from "react";

import { request } from "../utils/Fetch";
import { useAppContext } from "../AppContext";
import { NotificationType, useNotificationContext } from "./Notification";


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
    }, [pipeline]);

    return (
        <div className="flex flex-col fixed md:static w-full lg:w-2/3 h-screen overflow-y-scroll
            bg-stone-800 border-none lg:border-1 border-slate-900">
            { pipelineRuns.length < 1 &&
                <div className="text-3xl font-bold m-auto">
                    No Pipeline runs
                </div>
            }

            <div className="flex justify-between mt-8 ml-4 mr-6 text-2xl justify-self-end">
                <button onClick={() => { if (goBack) goBack(); }}
                    className="rounded-md px-4 py-2 bg-yellow-600 block md:hidden">&lt;- go back</button>
            </div>
        </div>
    );
};

export default DetailsPanel;