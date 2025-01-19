from flask import Flask, Response
import cv2
import numpy as np
from tflite_runtime.interpreter import Interpreter

app = Flask(__name__)

# Initialize TensorFlow Lite model once
interpreter = Interpreter('detect.tflite')
interpreter.allocate_tensors()

def set_input_tensor(interpreter, image):
    tensor_index = interpreter.get_input_details()[0]['index']
    input_tensor = interpreter.tensor(tensor_index)()[0]
    input_tensor[:, :] = np.expand_dims((image - 255) / 255, axis=0)

def get_output_tensor(interpreter, index):
    output_details = interpreter.get_output_details()[index]
    return np.squeeze(interpreter.get_tensor(output_details['index']))

def detect_objects(interpreter, image, threshold=0.6):
    set_input_tensor(interpreter, image)
    interpreter.invoke()
    boxes = get_output_tensor(interpreter, 1)
    classes = get_output_tensor(interpreter, 3)
    scores = get_output_tensor(interpreter, 0)
    count = int(get_output_tensor(interpreter, 2))
    return [
        {
            'bounding_box': boxes[i].tolist(),
            'class_id': classes[i].tolist(),
            'score': scores[i].tolist()
        }
        for i in range(count) if scores[i] >= threshold
    ]

def process_frame_with_detection(frame):
    # Resize and prepare frame for model
    resized_frame = cv2.resize(frame, (640, 640))
    
    # Run object detection
    results = detect_objects(interpreter, resized_frame)
    
    # Draw detection results on the frame
    for result in results:
        ymin, xmin, ymax, xmax = result['bounding_box']
        ymin, ymax = int(ymin * resized_frame.shape[0]), int(ymax * resized_frame.shape[0])
        xmin, xmax = int(xmin * resized_frame.shape[1]), int(xmax * resized_frame.shape[1])
        cv2.rectangle(resized_frame, (xmin, ymin), (xmax, ymax), (255, 0, 255), 2)
    return resized_frame

def generate_video_stream():
    video_path = "/path/to/your/video.mp4"
    cap = cv2.VideoCapture(video_path)
    
    while cap.isOpened():
        ret, frame = cap.read()
        if not ret:
            break
        
        # Apply object detection processing
        processed_frame = process_frame_with_detection(frame)
        
        # Encode the frame and send it to the client
        _, buffer = cv2.imencode('.jpg', processed_frame)
        yield (b'--frame\r\n'
               b'Content-Type: image/jpeg\r\n\r\n' + buffer.tobytes() + b'\r\n')

    cap.release()

@app.route('/process_video_stream')
def process_video_stream():
    return Response(generate_video_stream(), mimetype='multipart/x-mixed-replace; boundary=frame')

if __name__ == "__main__":
    app.run(debug=True)
