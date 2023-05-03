import React, { useState } from "react";
import axios from "axios";
import "./App.css";

const App = () => {
  const [message1, setMessage1] = useState("");
  const [message2, setMessage2] = useState("");

  const [message3, setMessage3] = useState("");
  const [result, setResult] = useState("");

  const handleInputChange1 = (event: React.ChangeEvent<HTMLInputElement>) => {
    setMessage1(event.target.value);
  };

  const handleInputChange2 = (event: React.ChangeEvent<HTMLInputElement>) => {
    setMessage2(event.target.value);
  };

  const handleInputChange3 = (event: React.ChangeEvent<HTMLInputElement>) => {
    setMessage3(event.target.value);
  };

  const handleButtonClicksend = async () => {
    await axios.get(`http://0.0.0.0:8080/api/send/${message1}/${message2}`);
    setMessage1("");
    setMessage2("");
  };

  const handleButtonClicksee = async () => {
    const response = await axios.get(`http://0.0.0.0:8080/api/see/${message3}/null`);
    setResult(response.data.message);
    setMessage3("");
  };

  return (
    <div className="container">
      <div className="input-group">
        <input type="text" value={message1} onChange={handleInputChange1} />
        <input type="text" value={message2} onChange={handleInputChange2} />
        <button className="button" onClick={handleButtonClicksend}>SEND</button>
      </div>
      <div className="input-group">
        <input type="text" value={message3} onChange={handleInputChange3} />
        <button className="button" onClick={handleButtonClicksee}>SEE</button>
        <h1>SECRET : {result}</h1>
      </div>
    </div>
  );

};

export default App;
