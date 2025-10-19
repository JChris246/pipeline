const SkipIcon = ({ className = "w-4 h-4", ...props }) => (
    <svg className={className} fill="none" stroke="currentColor" viewBox="0 0 24 24" {...props}>
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 4l10 8-10 8V4zM19 5v14" />
    </svg>
);

export default SkipIcon;
