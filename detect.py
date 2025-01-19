import cv2
from tflite_runtime.interpreter import Interpreter
import numpy as np

def set_input_tensor(interpreter, image):
    """Sets the input tensor with a preprocessed image."""
    tensor_index = interpreter.get_input_details()[0]['index']
    input_tensor = interpreter.tensor(tensor_index)()[0]
    input_tensor[:, :] = np.expand_dims((image - 255) / 255, axis=0)

def get_output_tensor(interpreter, index):
    """Returns the output tensor at the given index."""
    output_details = interpreter.get_output_details()[index]
    return np.squeeze(interpreter.get_tensor(output_details['index']))

def detect_objects(interpreter, image, threshold):
    """Detects objects in the image and returns detection results."""
    set_input_tensor(interpreter, image)
    interpreter.invoke()
    
    boxes = get_output_tensor(interpreter, 1)
    classes = get_output_tensor(interpreter, 3)
    scores = get_output_tensor(interpreter, 0)
    count = int(get_output_tensor(interpreter, 2))

    return [
        {
            'bounding_box': boxes[i],
            'class_id': classes[i],
            'score': scores[i]
        }
        for i in range(count) if scores[i] >= threshold
    ]

def process_image(img, interpreter, threshold):
    """Processes the image, detects objects, and crops the last detected license plate."""
    img_resized = cv2.resize(img, (640, 640))
    results = detect_objects(interpreter, img_resized, threshold)

    if results:
        # Crop the last detected license plate without drawing anything
        last_result = results[-1]
        ymin, xmin, ymax, xmax = last_result['bounding_box']
        xmin, xmax = int(max(1, xmin * img_resized.shape[1])), int(min(img_resized.shape[1], xmax * img_resized.shape[1]))
        ymin, ymax = int(max(1, ymin * img_resized.shape[0])), int(min(img_resized.shape[0], ymax * img_resized.shape[0]))

        # Crop the license plate region
        cv2.rectangle(img_resized, (xmin, ymin), (xmax, ymax), color=(255, 0, 255), thickness=2)

    return img_resized
def main():
    camera_link = "a09.mp4"  
    cap = cv2.VideoCapture(camera_link)
    
    # Check if video capture is successful
    if not cap.isOpened():
        print("Error: Couldn't open video.")
        return
    
    interpreter = Interpreter('detect.tflite')
    interpreter.allocate_tensors()

    threshold = 0.6  # Move threshold to outside loop to avoid redundant assignment
    while cap.isOpened():
        ret, img = cap.read()
        if not ret:
            break  # If reading fails, exit the loop
        
        img = process_image(img, interpreter, threshold)
        
        # Display the original frame with license plate detection
        cv2.imshow('Original Frame', img)
        
        # Exit the loop if 'q' is pressed
        if cv2.waitKey(1) & 0xFF == ord('q'):
            break

    cap.release()  # Release the video capture
    cv2.destroyAllWindows()  # Close all OpenCV windows

if __name__ == "__main__":
    main()
